import request from '@/api/request'
import type { ApiResult } from '@/api/request'
import type { UploadVideoRes } from '@/api/course'

// 分片上传 + 断点续传
//
// 流程：init（查已传分片，支持续传）→ 并发上传缺失分片 → merge（合并 + 后处理）
// uploadId 由文件元数据（name|size|lastModified|chunkSize）确定性生成，
// 同一文件刷新/重开浏览器后得到同一 id，服务端按 admin 命名空间隔离分片目录。

const CHUNK_SIZE = 5 * 1024 * 1024 // 5MB
const CONCURRENCY = 3

export interface ChunkProgress {
  phase: 'uploading' | 'merging'
  percent: number // 0-100，基于已传字节 / 文件总大小
  uploadedBytes: number
  totalBytes: number
}

export interface ChunkedUploadOptions {
  type: 'video' | 'courseware'
  thumbnail?: File | null // 仅 video，随 merge 一并提交
  onProgress?: (p: ChunkProgress) => void
  signal?: AbortSignal // 取消：abort 后停止调度 + 中断在传请求，merge 不调用
}

// cyrb53 轻量同步哈希（无需全文件哈希，零延迟）
function cyrb53(str: string, seed = 0): string {
  let h1 = 0xdeadbeef ^ seed
  let h2 = 0x41c6ce57 ^ seed
  for (let i = 0; i < str.length; i++) {
    const ch = str.charCodeAt(i)
    h1 = Math.imul(h1 ^ ch, 2654435761)
    h2 = Math.imul(h2 ^ ch, 1597334677)
  }
  h1 = Math.imul(h1 ^ (h1 >>> 16), 2246822507) ^ Math.imul(h2 ^ (h2 >>> 13), 3266489909)
  h2 = Math.imul(h2 ^ (h2 >>> 16), 2246822507) ^ Math.imul(h1 ^ (h1 >>> 13), 3266489909)
  const out = 4294967296 * (2097151 & h2) + (h2 >>> 0)
  return out.toString(36)
}

// 从文件元数据确定性生成 uploadId（刷新/重开浏览器后同一文件得到同一 id）
function computeUploadId(file: File): string {
  return cyrb53(`${file.name}|${file.size}|${file.lastModified}|${CHUNK_SIZE}`)
}

function chunkByteSize(index: number, fileSize: number): number {
  const start = index * CHUNK_SIZE
  return Math.min(CHUNK_SIZE, Math.max(0, fileSize - start))
}

/**
 * 分片上传一个文件（视频或课件），支持断点续传。
 * 返回合并后的 UploadVideoRes（按 type 填充对应字段）。
 */
export async function uploadFileChunked(
  file: File,
  opts: ChunkedUploadOptions
): Promise<UploadVideoRes> {
  const uploadId = computeUploadId(file)
  const totalChunks = Math.max(1, Math.ceil(file.size / CHUNK_SIZE))

  // 1. init：查已传分片
  const initRes = await request.post<
    any,
    ApiResult<{ uploaded: number[]; totalChunks: number }>
  >('/upload/chunk/init', {
    uploadId,
    filename: file.name,
    size: file.size,
    totalChunks,
    chunkSize: CHUNK_SIZE,
    type: opts.type
  })
  const uploadedSet = new Set(initRes.data.uploaded)
  const pending: number[] = []
  for (let i = 0; i < totalChunks; i++) {
    if (!uploadedSet.has(i)) pending.push(i)
  }

  // 已传字节（续传基线）
  let uploadedBytes = 0
  uploadedSet.forEach((i) => {
    uploadedBytes += chunkByteSize(i, file.size)
  })

  const report = (phase: 'uploading' | 'merging') => {
    opts.onProgress?.({
      phase,
      percent: Math.min(100, Math.round((uploadedBytes / file.size) * 100)),
      uploadedBytes,
      totalBytes: file.size
    })
  }
  report('uploading')

  // 2. 并发上传缺失分片
  let cursor = 0
  const worker = async () => {
    while (cursor < pending.length) {
      if (opts.signal?.aborted) throw new DOMException('Aborted', 'AbortError')
      const idx = pending[cursor++]
      const start = idx * CHUNK_SIZE
      const end = Math.min(start + CHUNK_SIZE, file.size)
      const blob = file.slice(start, end)
      const form = new FormData()
      form.append('uploadId', uploadId)
      form.append('chunkIndex', String(idx))
      form.append('file', blob, file.name)

      let chunkLoaded = 0
      const thisChunkSize = end - start
      await request.post('/upload/chunk', form, {
        timeout: 0,
        signal: opts.signal,
        onUploadProgress: (e) => {
          const newLoaded = e.loaded ?? 0
          // e.loaded 是本片累计已传字节，按增量计入全局
          uploadedBytes += newLoaded - chunkLoaded
          chunkLoaded = newLoaded
          report('uploading')
        }
      })
      // 兜底：确保本片完整计入（防止最后一次 progress 未触发）
      if (chunkLoaded < thisChunkSize) {
        uploadedBytes += thisChunkSize - chunkLoaded
        chunkLoaded = thisChunkSize
        report('uploading')
      }
    }
  }
  const workers = Array.from(
    { length: Math.min(CONCURRENCY, pending.length) },
    () => worker()
  )
  await Promise.all(workers)

  // 3. merge：流式合并 + 后处理（ffprobe/LibreOffice，可能耗时数分钟）
  uploadedBytes = file.size
  opts.onProgress?.({
    phase: 'merging',
    percent: 100,
    uploadedBytes,
    totalBytes: file.size
  })
  const mergeForm = new FormData()
  mergeForm.append('uploadId', uploadId)
  mergeForm.append('filename', file.name)
  mergeForm.append('type', opts.type)
  mergeForm.append('totalChunks', String(totalChunks))
  mergeForm.append('size', String(file.size))
  if (opts.type === 'video' && opts.thumbnail) {
    mergeForm.append('thumbnail', opts.thumbnail)
  }
  const mergeRes = await request.post<any, ApiResult<UploadVideoRes>>(
    '/upload/chunk/merge',
    mergeForm,
    { timeout: 0, signal: opts.signal }
  )
  return mergeRes.data
}
