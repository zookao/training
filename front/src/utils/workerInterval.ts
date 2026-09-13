/**
 * 免受后台标签页节流的周期定时器（Web Worker 驱动）。
 *
 * 背景：Chrome/Edge 会节流后台标签页主线程的 setInterval（隐藏超 5 分钟后
 * 降至约每分钟 1 次）。学习页的 10 秒进度上报一旦被节流，服务端每次上报
 * 最多只承认 13 秒前进（防快进上限），挂后台听课的学员每真实观看 1 分钟
 * 只被记 13 秒，进度大量丢失且永远无法完课。Web Worker 内的定时器不受
 * 该节流策略约束，故用 Worker 按周期 postMessage 唤醒主线程执行回调。
 *
 * Worker 代码内联为 Blob，无需构建配置配合；环境不支持（老旧浏览器 /
 * CSP 禁止 blob: worker）时回退普通 setInterval，行为退化为修复前。
 */

export interface StableInterval {
  /** 停止定时器并释放 Worker 资源，可安全重复调用 */
  clear: () => void
}

// Worker 源码必须是纯 JS（浏览器不执行 TS）
const WORKER_SRC = `
let timer = null
self.onmessage = (e) => {
  const cmd = e.data && e.data.cmd
  if (cmd === 'start') {
    if (timer !== null) clearInterval(timer)
    timer = setInterval(() => self.postMessage('tick'), e.data.ms)
  } else if (cmd === 'stop') {
    if (timer !== null) clearInterval(timer)
    close()
  }
}
`

export function setWorkerInterval(callback: () => void, ms: number): StableInterval {
  try {
    const url = URL.createObjectURL(new Blob([WORKER_SRC], { type: 'application/javascript' }))
    const worker = new Worker(url)
    worker.onmessage = () => callback()
    worker.postMessage({ cmd: 'start', ms })
    return {
      clear() {
        try {
          worker.postMessage({ cmd: 'stop' })
          worker.terminate()
        } finally {
          URL.revokeObjectURL(url)
        }
      }
    }
  } catch {
    const id = window.setInterval(callback, ms)
    return { clear: () => window.clearInterval(id) }
  }
}
