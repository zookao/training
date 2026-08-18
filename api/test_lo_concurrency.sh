#!/bin/bash
# ===================================================================
# LibreOffice 并发转换排队 & 超时机制验证脚本
# 模拟 10 个用户同时上传课件（PPTX），验证：
#   1. 排队机制：maxConcurrency=1 时，10 个转换串行执行，全部成功
#   2. 超时机制：queueTimeoutSec=2 时，排队的请求超时返回错误
# ===================================================================

set -uo pipefail

BASE="http://127.0.0.1:8000"
ADMIN_API="$BASE/api/admin"
API_DIR="/Users/zookao/goCode/self/training/api"
SRC_PPTX="$API_DIR/upload/260807/pptx/e4294492f2a959b2aaaea00a72d45fbb.pptx"
CONFIG_FILE="$API_DIR/config.yaml"
BACKEND_LOG="/tmp/training_backend.log"
NUM_USERS=10
TMPDIR=$(mktemp -d)
PASS=0
FAIL=0

post() { curl -s -X POST "$1" -H "Content-Type: application/json" -H "Authorization: Bearer $2" -d "$3"; }
pass() { PASS=$((PASS+1)); echo "  ✅ $1"; }
fail() { FAIL=$((FAIL+1)); echo "  ❌ $1 | $2"; }

# ---- 辅助函数 ----

# 重启后端
restart_backend() {
  local pid=$(pgrep -f './training' | head -1)
  [ -n "$pid" ] && kill "$pid" 2>/dev/null || true
  sleep 1
  cd "$API_DIR" && nohup ./training > "$BACKEND_LOG" 2>&1 &
  sleep 2
  # 等待启动
  for i in $(seq 1 10); do
    if curl -s "$BASE/api/admin/auth/login" -X POST -H "Content-Type: application/json" \
       -d '{"username":"admin","password":"admin123"}' | jq -e '.code==0' >/dev/null 2>&1; then
      return 0
    fi
    sleep 1
  done
  echo "❌ 后端启动失败"; exit 1
}

# 修改 config.yaml 中的 queueTimeoutSec
set_queue_timeout() {
  local sec=$1
  cd "$API_DIR"
  if grep -q 'queueTimeoutSec' "$CONFIG_FILE"; then
    sed -i '' "s/queueTimeoutSec:.*/queueTimeoutSec: $sec/" "$CONFIG_FILE"
  fi
}

# 获取文件大小（字节）
FILE_SIZE=$(stat -f%z "$SRC_PPTX" 2>/dev/null || stat -c%s "$SRC_PPTX" 2>/dev/null)
echo "源 PPTX: $SRC_PPTX ($FILE_SIZE bytes)"
echo ""

# 预处理：为每个用户复制一份 PPTX（不同文件名 → 不同 hash → 不同输出路径）
for i in $(seq 1 $NUM_USERS); do
  cp "$SRC_PPTX" "$TMPDIR/user_${i}.pptx"
done

# ===================================================================
# Part 1: 排队机制验证（正常超时 1800s，maxConcurrency=1）
# ===================================================================
echo "==================================================================="
echo "  Part 1: 排队机制验证 (maxConcurrency=1, queueTimeout=1800s)"
echo "==================================================================="
echo ""

# 确保正常配置
set_queue_timeout 1800
restart_backend
ADMIN_TOKEN=$(post "$ADMIN_API/auth/login" "" '{"username":"admin","password":"admin123"}' | jq -r '.data.token')
echo "管理员登录成功，maxConcurrency=1"

# 预分片：为每个用户 init + upload chunk（串行，快速完成，不触发转换）
echo ""
echo "【预处理】为 $NUM_USERS 个用户初始化上传会话并上传分片..."
for i in $(seq 1 $NUM_USERS); do
  UPLOAD_ID="lo_test_${i}_$(date +%s)"
  FILENAME="lo_concurrent_${i}.pptx"
  # init
  post "$ADMIN_API/upload/chunk/init" "$ADMIN_TOKEN" \
    "{\"uploadId\":\"$UPLOAD_ID\",\"filename\":\"$FILENAME\",\"size\":$FILE_SIZE,\"totalChunks\":1,\"chunkSize\":$FILE_SIZE,\"type\":\"courseware\"}" > /dev/null
  # upload chunk
  curl -s -X POST "$ADMIN_API/upload/chunk" \
    -H "Authorization: Bearer $ADMIN_TOKEN" \
    -F "uploadId=$UPLOAD_ID" \
    -F "chunkIndex=0" \
    -F "file=@$TMPDIR/user_${i}.pptx" > /dev/null
  # 保存信息供并发 merge 使用
  echo "$UPLOAD_ID $FILENAME" >> "$TMPDIR/uploads.txt"
done
echo "  ✅ $NUM_USERS 个用户分片上传完成"
echo ""

# 并发 merge：10 个请求同时触发 ConvertToPDF
echo "【并发交卷】$NUM_USERS 个用户同时提交 merge 请求（触发 LibreOffice 转换）..."
GLOBAL_START=$(python3 -c "import time; print(time.time())")

PIDS=()
for i in $(seq 1 $NUM_USERS); do
  read -r UPLOAD_ID FILENAME <<< $(sed -n "${i}p" "$TMPDIR/uploads.txt")
  (
    START=$(python3 -c "import time; print(time.time())")
    RESP=$(curl -s -X POST "$ADMIN_API/upload/chunk/merge" \
      -H "Authorization: Bearer $ADMIN_TOKEN" \
      -F "uploadId=$UPLOAD_ID" \
      -F "filename=$FILENAME" \
      -F "type=courseware" \
      -F "totalChunks=1" \
      -F "size=$FILE_SIZE")
    END=$(python3 -c "import time; print(time.time())")
    ELAPSED=$(python3 -c "print(f'{$END-$START:.1f}')")
    CODE=$(echo "$RESP" | jq -r '.code // "null"')
    PDF_URL=$(echo "$RESP" | jq -r '.data.coursewarePdf // "?"')
    MSG=$(echo "$RESP" | jq -r '.msg // "?"')
    echo "$i $ELAPSED $CODE $PDF_URL $MSG" > "$TMPDIR/result_${i}.txt"
  ) &
  PIDS+=($!)
done
for pid in "${PIDS[@]}"; do wait $pid; done

GLOBAL_END=$(python3 -c "import time; print(time.time())")
TOTAL_ELAPSED=$(python3 -c "print(f'{$GLOBAL_END-$GLOBAL_START:.1f}')")

# 分析结果
echo ""
echo "【结果分析】"
echo "  用户  耗时(s)  状态    PDF路径"
echo "  ----  -------  ------  --------"
SUCCESS=0
FAILED=0
TIMES=""
for i in $(seq 1 $NUM_USERS); do
  LINE=$(cat "$TMPDIR/result_${i}.txt")
  UIDX=$(echo "$LINE" | awk '{print $1}')
  ELAPSED=$(echo "$LINE" | awk '{print $2}')
  CODE=$(echo "$LINE" | awk '{print $3}')
  PDF_URL=$(echo "$LINE" | awk '{print $4}')
  if [ "$CODE" = "0" ]; then
    STATUS="✅成功"
    SUCCESS=$((SUCCESS+1))
    TIMES="$TIMES $ELAPSED"
  else
    STATUS="❌失败"
    FAILED=$((FAILED+1))
    TIMES="$TIMES $ELAPSED"
  fi
  printf "  %3d    %5s    %s  %s\n" "$i" "$ELAPSED" "$STATUS" "$PDF_URL"
done

echo ""
echo "  总耗时: ${TOTAL_ELAPSED}s | 成功: $SUCCESS | 失败: $FAILED"
echo "  各请求耗时(s):$TIMES"

# 验证排队：如果串行执行，总耗时 ≈ 单次耗时 × 10
SINGLE_TIME=$(echo "$TIMES" | tr ' ' '\n' | grep -v '^$' | sort -n | head -1)
LAST_TIME=$(echo "$TIMES" | tr ' ' '\n' | grep -v '^$' | sort -n | tail -1)
echo "  最快: ${SINGLE_TIME}s | 最慢: ${LAST_TIME}s"

if [ "$SUCCESS" -eq "$NUM_USERS" ]; then
  pass "排队机制: $NUM_USERS 个并发上传全部成功（无转换失败）"
else
  fail "排队机制: 有 $FAILED 个上传失败" "成功 $SUCCESS/$NUM_USERS"
fi

# 验证串行：最慢的应该明显比最快的慢（排队等待）
if python3 -c "exit(0 if float('$LAST_TIME') > float('$SINGLE_TIME') * 2 else 1)" 2>/dev/null; then
  pass "串行排队: 最慢(${LAST_TIME}s) > 2×最快(${SINGLE_TIME}s)，排队生效"
else
  fail "串行排队: 排队不明显" "最快=${SINGLE_TIME}s, 最慢=${LAST_TIME}s"
fi

# 检查后端日志中的排队信息
LO_LOG_LINES=$(grep '\[LibreOffice\]' "$BACKEND_LOG" | tail -30)
LO_START_COUNT=$(echo "$LO_LOG_LINES" | grep '开始转换' | wc -l | tr -d ' ')
LO_DONE_COUNT=$(echo "$LO_LOG_LINES" | grep '转换完成' | wc -l | tr -d ' ')
echo ""
echo "  后端日志: 转换开始 ${LO_START_COUNT} 次, 转换完成 ${LO_DONE_COUNT} 次"
if [ "$LO_START_COUNT" -gt 0 ]; then
  pass "日志可观测: 排队日志正常输出"
else
  fail "日志可观测: 未找到排队日志" "检查 $BACKEND_LOG"
fi
echo ""

# ===================================================================
# Part 2: 超时机制验证（queueTimeoutSec=2，maxConcurrency=1）
# ===================================================================
echo "==================================================================="
echo "  Part 2: 超时机制验证 (maxConcurrency=1, queueTimeoutSec=2)"
echo "==================================================================="
echo ""

# 修改超时为 2 秒，重启后端
echo "【修改配置】queueTimeoutSec: 2 → 重启后端..."
set_queue_timeout 2
restart_backend
ADMIN_TOKEN=$(post "$ADMIN_API/auth/login" "" '{"username":"admin","password":"admin123"}' | jq -r '.data.token')
echo "  后端已重启，排队超时 = 2s"
echo ""

# 清空日志
> "$BACKEND_LOG"

# 预分片
echo "【预处理】为 $NUM_USERS 个用户初始化上传会话..."
rm -f "$TMPDIR/uploads2.txt"
for i in $(seq 1 $NUM_USERS); do
  UPLOAD_ID="lo_timeout_${i}_$(date +%s)"
  FILENAME="lo_timeout_${i}.pptx"
  post "$ADMIN_API/upload/chunk/init" "$ADMIN_TOKEN" \
    "{\"uploadId\":\"$UPLOAD_ID\",\"filename\":\"$FILENAME\",\"size\":$FILE_SIZE,\"totalChunks\":1,\"chunkSize\":$FILE_SIZE,\"type\":\"courseware\"}" > /dev/null
  curl -s -X POST "$ADMIN_API/upload/chunk" \
    -H "Authorization: Bearer $ADMIN_TOKEN" \
    -F "uploadId=$UPLOAD_ID" \
    -F "chunkIndex=0" \
    -F "file=@$TMPDIR/user_${i}.pptx" > /dev/null
  echo "$UPLOAD_ID $FILENAME" >> "$TMPDIR/uploads2.txt"
done
echo "  ✅ $NUM_USERS 个用户分片上传完成"
echo ""

# 并发 merge
echo "【并发交卷】$NUM_USERS 个用户同时提交 merge 请求..."
echo "  期望: 第 1 个成功（~4s 转换），其余排队超时失败（>2s 等待）"
echo ""

PIDS=()
for i in $(seq 1 $NUM_USERS); do
  read -r UPLOAD_ID FILENAME <<< $(sed -n "${i}p" "$TMPDIR/uploads2.txt")
  (
    START=$(python3 -c "import time; print(time.time())")
    RESP=$(curl -s -X POST "$ADMIN_API/upload/chunk/merge" \
      -H "Authorization: Bearer $ADMIN_TOKEN" \
      -F "uploadId=$UPLOAD_ID" \
      -F "filename=$FILENAME" \
      -F "type=courseware" \
      -F "totalChunks=1" \
      -F "size=$FILE_SIZE")
    END=$(python3 -c "import time; print(time.time())")
    ELAPSED=$(python3 -c "print(f'{$END-$START:.1f}')")
    CODE=$(echo "$RESP" | jq -r '.code // "null"')
    MSG=$(echo "$RESP" | jq -r '.msg // "?"')
    echo "$i $ELAPSED $CODE $MSG" > "$TMPDIR/timeout_result_${i}.txt"
  ) &
  PIDS+=($!)
done
for pid in "${PIDS[@]}"; do wait $pid; done

# 分析结果
echo "【结果分析】"
echo "  用户  耗时(s)  状态    消息"
echo "  ----  -------  ------  ----"
TIMEOUT_OK=0
TIMEOUT_FAIL=0
TIMEOUT_MSGS=0
for i in $(seq 1 $NUM_USERS); do
  LINE=$(cat "$TMPDIR/timeout_result_${i}.txt")
  UIDX=$(echo "$LINE" | awk '{print $1}')
  ELAPSED=$(echo "$LINE" | awk '{print $2}')
  CODE=$(echo "$LINE" | awk '{print $3}')
  MSG=$(echo "$LINE" | cut -d' ' -f4-)
  if [ "$CODE" = "0" ]; then
    STATUS="✅成功"
    TIMEOUT_OK=$((TIMEOUT_OK+1))
  else
    STATUS="❌失败"
    TIMEOUT_FAIL=$((TIMEOUT_FAIL+1))
    # 检查是否是超时错误
    if echo "$MSG" | grep -q '超时\|排队'; then
      TIMEOUT_MSGS=$((TIMEOUT_MSGS+1))
    fi
  fi
  printf "  %3d    %5s    %s  %s\n" "$i" "$ELAPSED" "$STATUS" "$MSG"
done

echo ""
echo "  成功: $TIMEOUT_OK | 失败: $TIMEOUT_FAIL | 其中超时错误: $TIMEOUT_MSGS"

# 验证：应该有大部分请求超时失败
if [ "$TIMEOUT_OK" -ge 1 ] && [ "$TIMEOUT_FAIL" -ge 1 ]; then
  pass "超时机制: $TIMEOUT_OK 个成功, $TIMEOUT_FAIL 个超时失败"
else
  fail "超时机制: 结果不符合预期" "成功=$TIMEOUT_OK, 失败=$TIMEOUT_FAIL"
fi

if [ "$TIMEOUT_MSGS" -ge 1 ]; then
  pass "超时错误提示: $TIMEOUT_MSGS 个请求返回了排队超时提示"
else
  fail "超时错误提示: 未返回排队超时提示" "检查失败请求的 msg"
fi
echo ""

# ===================================================================
# 恢复配置
# ===================================================================
echo "==================================================================="
echo "  恢复配置"
echo "==================================================================="
set_queue_timeout 1800
restart_backend
echo "  ✅ 已恢复 queueTimeoutSec: 1800，后端已重启"

# 清理
rm -rf "$TMPDIR"

# 汇总
echo ""
echo "==================================================================="
echo "  测试汇总"
echo "==================================================================="
echo "  通过: $PASS"
echo "  失败: $FAIL"
echo "  总计: $((PASS+FAIL))"
echo ""
if [ $FAIL -eq 0 ]; then
  echo "  🎉 全部通过！排队和超时机制均正常工作"
else
  echo "  ⚠️  有 $FAIL 个用例未通过"
fi
exit $FAIL
