#!/bin/bash
# ===================================================================
# 培训系统并发测试脚本
# 模拟多用户同时考试/上报进度，检测竞态条件与数据一致性
# ===================================================================

BASE="http://127.0.0.1:8000"
ADMIN_API="$BASE/api/admin"
USER_API="$BASE/api/user"
TS=$(date +%s)
COURSE_ID=13
CLASS_ID=11
VIDEO_ID=8
VIDEO_DURATION=20
TESTPAPER_ID=26
Q1_ID=31
Q2_ID=32
NUM_STUDENTS=10
PASS=0
FAIL=0
FAILED_CASES=()
TMPDIR=$(mktemp -d)

post() { curl -s -X POST "$1" -H "Content-Type: application/json" -H "Authorization: Bearer $2" -d "$3"; }
get()  { curl -s -X GET "$1" -H "Authorization: Bearer $2"; }
put()  { curl -s -X PUT "$1" -H "Content-Type: application/json" -H "Authorization: Bearer $2" -d "$3"; }
del()  { curl -s -X DELETE "$1" -H "Authorization: Bearer $2"; }

pass() { PASS=$((PASS+1)); echo "  ✅ $1"; }
fail() { FAIL=$((FAIL+1)); FAILED_CASES+=("$1 → $2"); echo "  ❌ $1 | $2"; }

echo "==================================================================="
echo "  培训系统并发测试  (TS=$TS, students=$NUM_STUDENTS)"
echo "==================================================================="
echo ""

# ==== 0. 管理员登录 ====
ADMIN_TOKEN=$(post "$ADMIN_API/auth/login" "" '{"username":"admin","password":"admin123"}' | jq -r '.data.token')
[ -z "$ADMIN_TOKEN" ] && { echo "❌ 管理员登录失败"; exit 1; }
echo "管理员登录成功"

# ==== 0.5 清理上次残留的 conc_ 测试学员 ====
LEFTOVER_IDS=$(get "$ADMIN_API/user?page=1&pageSize=500" "$ADMIN_TOKEN" | jq -r '.data.list[] | select(.username | startswith("conc_")) | .id' | tr '\n' ' ')
if [ -n "$LEFTOVER_IDS" ]; then
  for uid in $LEFTOVER_IDS; do
    del "$ADMIN_API/user/$uid" "$ADMIN_TOKEN" > /dev/null 2>&1
  done
  echo "清理上次残留测试学员: ${LEFTOVER_IDS}"
fi

# ==== 1. 创建测试学员 ====
echo ""
echo "【1. 创建 $NUM_STUDENTS 个测试学员】"
STUDENT_USERS=()
STUDENT_TOKENS=()
STUDENT_IDS=()
for i in $(seq 1 $NUM_STUDENTS); do
  SUSER="conc_${TS}_$i"
  PHONE="14${TS: -7}$(printf '%02d' $i)"
  BODY='{"username":"'"$SUSER"'","password":"123456","nickname":"并发学员'"$i"'","phone":"'"$PHONE"'","status":1}'
  post "$ADMIN_API/user" "$ADMIN_TOKEN" "$BODY" > /dev/null
  SID=$(get "$ADMIN_API/user?page=1&pageSize=500" "$ADMIN_TOKEN" | jq -r ".data.list[] | select(.username==\"$SUSER\") | .id")
  # 学员登录
  STU_TOKEN=$(post "$USER_API/auth/login" "" '{"username":"'"$SUSER"'","password":"123456"}' | jq -r '.data.token')
  STUDENT_USERS+=("$SUSER")
  STUDENT_TOKENS+=("$STU_TOKEN")
  STUDENT_IDS+=("$SID")
done
# 统一分配到班级：PUT /class/:id/users 是全量替换语义，必须一次性传入所有学员 ID
ALL_IDS=$(IFS=,; echo "${STUDENT_IDS[*]}")
ASSIGN_RESP=$(put "$ADMIN_API/class/$CLASS_ID/users" "$ADMIN_TOKEN" '{"ids":['"$ALL_IDS"']}')
ASSIGN_CODE=$(echo "$ASSIGN_RESP" | jq -r '.code')
if [ "$ASSIGN_CODE" = "0" ]; then
  echo "  ✅ 创建 $NUM_STUDENTS 个学员并统一分配到班级"
else
  echo "  ❌ 班级分配失败: $ASSIGN_RESP"
fi
echo ""

# ===================================================================
# 场景 1: 同一学员并发跳跃上报（防作弊竞态）
# 目标：20个并发请求同时上报 position=600，验证 maxPosition 被限速
# ===================================================================
echo "【场景1: 同一学员并发跳跃上报（防作弊竞态）】"
JUMP_USER="${STUDENT_USERS[0]}"
JUMP_TOKEN="${STUDENT_TOKENS[0]}"
JUMP_ID="${STUDENT_IDS[0]}"

# 清除该学员的已有进度记录（通过 DB）
docker exec mysql80 mysql -uroot -proot training -e "DELETE FROM video_records WHERE user_id=$JUMP_ID AND video_id=$VIDEO_ID;" 2>/dev/null
docker exec mysql80 mysql -uroot -proot training -e "DELETE FROM check_logs WHERE user_id=$JUMP_ID AND video_id=$VIDEO_ID;" 2>/dev/null

echo "  学员: $JUMP_USER (id=$JUMP_ID)"
echo "  发送 20 个并发请求 position=600 (期望被限速到 ~13s)..."

# 并发发送 20 个跳跃上报
PIDS=()
for i in $(seq 1 20); do
  BODY='{"videoId":'"$VIDEO_ID"',"courseId":'"$COURSE_ID"',"classId":'"$CLASS_ID"',"position":600,"duration":'"$VIDEO_DURATION"',"completed":true}'
  post "$USER_API/progress" "$JUMP_TOKEN" "$BODY" > "$TMPDIR/jump_$i.json" &
  PIDS+=($!)
done
for pid in "${PIDS[@]}"; do wait $pid; done

# 检查结果
CONC_OK=0
for f in "$TMPDIR"/jump_*.json; do
  [ "$(jq -r '.code // "null"' "$f" 2>/dev/null)" = "0" ] && CONC_OK=$((CONC_OK+1))
done
echo "  并发请求成功数: $CONC_OK / 20"

# 查 DB 验证 maxPosition
MAX_POS=$(docker exec mysql80 mysql -uroot -proot training -N -e "SELECT max_position FROM video_records WHERE user_id=$JUMP_ID AND video_id=$VIDEO_ID;" 2>/dev/null)
COMPLETED=$(docker exec mysql80 mysql -uroot -proot training -N -e "SELECT completed FROM video_records WHERE user_id=$JUMP_ID AND video_id=$VIDEO_ID;" 2>/dev/null)
CHECK_PENDING=$(docker exec mysql80 mysql -uroot -proot training -N -e "SELECT check_pending FROM video_records WHERE user_id=$JUMP_ID AND video_id=$VIDEO_ID;" 2>/dev/null)
echo "  DB 结果: maxPosition=$MAX_POS, completed=$COMPLETED, checkPending=$CHECK_PENDING"

if [ -n "$MAX_POS" ] && [ "$MAX_POS" -lt 100 ] 2>/dev/null; then
  pass "并发跳跃被限速 (maxPos=$MAX_POS < 100)"
else
  fail "并发跳跃未限速" "maxPos=$MAX_POS (期望 <100)"
fi

if [ "$COMPLETED" = "0" ]; then
  pass "并发跳跃未标记完成"
else
  fail "并发跳跃被标记完成" "completed=$COMPLETED"
fi

if [ "$CHECK_PENDING" = "1" ]; then
  pass "并发跳跃触发校验拦截 (checkPending=1)"
else
  fail "并发跳跃未触发校验" "checkPending=$CHECK_PENDING"
fi
echo ""

# ===================================================================
# 场景 2: 多学员并发上报进度
# 目标：10个学员同时上报进度，验证全部成功无错误
# ===================================================================
echo "【场景2: $NUM_STUDENTS 个学员并发上报进度】"
PIDS=()
for i in $(seq 0 $((NUM_STUDENTS-1))); do
  TOKEN="${STUDENT_TOKENS[$i]}"
  BODY='{"videoId":'"$VIDEO_ID"',"courseId":'"$COURSE_ID"',"classId":'"$CLASS_ID"',"position":5,"duration":'"$VIDEO_DURATION"',"completed":false}'
  post "$USER_API/progress" "$TOKEN" "$BODY" > "$TMPDIR/multi_prog_$i.json" &
  PIDS+=($!)
done
for pid in "${PIDS[@]}"; do wait $pid; done

MULTI_OK=0
MULTI_FAIL=0
for i in $(seq 0 $((NUM_STUDENTS-1))); do
  CODE=$(jq -r '.code // "null"' "$TMPDIR/multi_prog_$i.json" 2>/dev/null)
  if [ "$CODE" = "0" ]; then MULTI_OK=$((MULTI_OK+1)); else MULTI_FAIL=$((MULTI_FAIL+1)); fi
done
echo "  成功: $MULTI_OK, 失败: $MULTI_FAIL"

if [ "$MULTI_OK" -eq "$NUM_STUDENTS" ]; then
  pass "$NUM_STUDENTS 个学员并发上报全部成功"
else
  fail "多学员并发上报有失败" "成功 $MULTI_OK / $NUM_STUDENTS"
fi

# 验证每个学员的 maxPosition 都被正确限速
ALL_CAPPED=true
for i in $(seq 0 $((NUM_STUDENTS-1))); do
  SID="${STUDENT_IDS[$i]}"
  MP=$(docker exec mysql80 mysql -uroot -proot training -N -e "SELECT max_position FROM video_records WHERE user_id=$SID AND video_id=$VIDEO_ID;" 2>/dev/null)
  if [ -n "$MP" ] && [ "$MP" -gt 20 ] 2>/dev/null; then
    ALL_CAPPED=false
    echo "  ⚠️ 学员 $SID maxPosition=$MP (异常)"
  fi
done
if $ALL_CAPPED; then
  pass "所有学员 maxPosition 均被正确限速"
else
  fail "部分学员 maxPosition 异常" "见上方警告"
fi
echo ""

# ===================================================================
# 场景 3: 并发进入考试（10个学员同时进入）
# ===================================================================
echo "【场景3: $NUM_STUDENTS 个学员并发进入考试】"
PIDS=()
for i in $(seq 0 $((NUM_STUDENTS-1))); do
  TOKEN="${STUDENT_TOKENS[$i]}"
  get "$USER_API/testpaper/$TESTPAPER_ID/exam" "$TOKEN" > "$TMPDIR/exam_enter_$i.json" &
  PIDS+=($!)
done
for pid in "${PIDS[@]}"; do wait $pid; done

EXAM_OK=0
EXAM_FAIL=0
for i in $(seq 0 $((NUM_STUDENTS-1))); do
  CODE=$(jq -r '.code // "null"' "$TMPDIR/exam_enter_$i.json" 2>/dev/null)
  QCOUNT=$(jq -r '.data.questions | length' "$TMPDIR/exam_enter_$i.json" 2>/dev/null)
  if [ "$CODE" = "0" ] && [ "$QCOUNT" = "2" ]; then EXAM_OK=$((EXAM_OK+1)); else EXAM_FAIL=$((EXAM_FAIL+1)); fi
done
echo "  成功进入: $EXAM_OK, 失败: $EXAM_FAIL"

if [ "$EXAM_OK" -eq "$NUM_STUDENTS" ]; then
  pass "$NUM_STUDENTS 个学员并发进入考试全部成功"
else
  fail "并发进入考试有失败" "成功 $EXAM_OK / $NUM_STUDENTS"
fi
echo ""

# ===================================================================
# 场景 4: 并发交卷（10个学员同时提交）
# ===================================================================
echo "【场景4: $NUM_STUDENTS 个学员并发交卷】"
PIDS=()
for i in $(seq 0 $((NUM_STUDENTS-1))); do
  TOKEN="${STUDENT_TOKENS[$i]}"
  BODY='{"answers":[{"questionId":'"$Q1_ID"',"userAnswer":["B"]},{"questionId":'"$Q2_ID"',"userAnswer":["A"]}]}'
  post "$USER_API/testpaper/$TESTPAPER_ID/submit" "$TOKEN" "$BODY" > "$TMPDIR/exam_submit_$i.json" &
  PIDS+=($!)
done
for pid in "${PIDS[@]}"; do wait $pid; done

SUBMIT_OK=0
SUBMIT_FAIL=0
SCORES=""
for i in $(seq 0 $((NUM_STUDENTS-1))); do
  CODE=$(jq -r '.code // "null"' "$TMPDIR/exam_submit_$i.json" 2>/dev/null)
  SCORE=$(jq -r '.data.score // "?"' "$TMPDIR/exam_submit_$i.json" 2>/dev/null)
  PASSED=$(jq -r '.data.passed // "?"' "$TMPDIR/exam_submit_$i.json" 2>/dev/null)
  if [ "$CODE" = "0" ]; then
    SUBMIT_OK=$((SUBMIT_OK+1))
    SCORES="$SCORES $SCORE"
  else
    SUBMIT_FAIL=$((SUBMIT_FAIL+1))
    echo "  ⚠️ 学员$i 交卷失败: $(cat "$TMPDIR/exam_submit_$i.json" | head -c 150)"
  fi
done
echo "  成功交卷: $SUBMIT_OK, 失败: $SUBMIT_FAIL"
echo "  分数: $SCORES"

if [ "$SUBMIT_OK" -eq "$NUM_STUDENTS" ]; then
  pass "$NUM_STUDENTS 个学员并发交卷全部成功"
else
  fail "并发交卷有失败" "成功 $SUBMIT_OK / $NUM_STUDENTS"
fi

# 验证所有分数都是100
ALL_100=true
for i in $(seq 0 $((NUM_STUDENTS-1))); do
  SCORE=$(jq -r '.data.score // "0"' "$TMPDIR/exam_submit_$i.json" 2>/dev/null)
  if [ "$SCORE" != "100" ]; then
    ALL_100=false
    echo "  ⚠️ 学员$i 分数=$SCORE (期望100)"
  fi
done
if $ALL_100; then
  pass "所有学员分数正确 (100分)"
else
  fail "部分学员分数异常" "见上方警告"
fi

# 验证 DB 中考试记录数 = 学员数（无重复）
RECORD_COUNT=$(docker exec mysql80 mysql -uroot -proot training -N -e "SELECT COUNT(*) FROM testpaper_records WHERE testpaper_id=$TESTPAPER_ID AND submitted_at IS NOT NULL AND user_id IN ($(IFS=,; echo "${STUDENT_IDS[*]}"));" 2>/dev/null)
echo "  DB 已交卷记录数: $RECORD_COUNT (期望 $NUM_STUDENTS)"
if [ "$RECORD_COUNT" = "$NUM_STUDENTS" ]; then
  pass "考试记录数正确无重复 ($RECORD_COUNT = $NUM_STUDENTS)"
else
  fail "考试记录数异常" "DB=$RECORD_COUNT, 期望=$NUM_STUDENTS"
fi
echo ""

# ===================================================================
# 场景 5: 同一学员并发进入考试（重复记录检测）
# 目标：同一学员5次并发进入考试，应只创建1条未提交记录
# ===================================================================
echo "【场景5: 同一学员并发进入考试（重复记录检测）】"
DUP_USER="${STUDENT_USERS[1]}"
DUP_TOKEN="${STUDENT_TOKENS[1]}"
DUP_ID="${STUDENT_IDS[1]}"

# 先清除该学员的考试记录
docker exec mysql80 mysql -uroot -proot training -e "DELETE FROM testpaper_records WHERE user_id=$DUP_ID AND testpaper_id=$TESTPAPER_ID;" 2>/dev/null

echo "  学员: $DUP_USER (id=$DUP_ID)"
echo "  发送 5 个并发进入考试请求..."

PIDS=()
for i in $(seq 1 5); do
  get "$USER_API/testpaper/$TESTPAPER_ID/exam" "$DUP_TOKEN" > "$TMPDIR/dup_enter_$i.json" &
  PIDS+=($!)
done
for pid in "${PIDS[@]}"; do wait $pid; done

DUP_OK=0
for i in $(seq 1 5); do
  CODE=$(jq -r '.code // "null"' "$TMPDIR/dup_enter_$i.json" 2>/dev/null)
  if [ "$CODE" = "0" ]; then DUP_OK=$((DUP_OK+1)); fi
done
echo "  成功响应: $DUP_OK / 5"

# 检查 DB 中未提交记录数
UNSUBMITTED=$(docker exec mysql80 mysql -uroot -proot training -N -e "SELECT COUNT(*) FROM testpaper_records WHERE user_id=$DUP_ID AND testpaper_id=$TESTPAPER_ID AND submitted_at IS NULL;" 2>/dev/null)
echo "  DB 未提交记录数: $UNSUBMITTED (期望 1)"

if [ "$UNSUBMITTED" = "1" ]; then
  pass "并发进入考试未产生重复记录 ($UNSUBMITTED = 1)"
else
  fail "并发进入考试产生重复记录" "DB未提交记录=$UNSUBMITTED (期望1)"
fi
echo ""

# ===================================================================
# 场景 6: 同一学员并发交卷（重复提交检测）
# ===================================================================
echo "【场景6: 同一学员并发交卷（重复提交检测）】"
DUP2_USER="${STUDENT_USERS[2]}"
DUP2_TOKEN="${STUDENT_TOKENS[2]}"
DUP2_ID="${STUDENT_IDS[2]}"

# 清除已有记录，先进入考试
docker exec mysql80 mysql -uroot -proot training -e "DELETE FROM testpaper_records WHERE user_id=$DUP2_ID AND testpaper_id=$TESTPAPER_ID;" 2>/dev/null
get "$USER_API/testpaper/$TESTPAPER_ID/exam" "$DUP2_TOKEN" > /dev/null

echo "  学员: $DUP2_USER (id=$DUP2_ID)"
echo "  发送 5 个并发交卷请求..."

PIDS=()
for i in $(seq 1 5); do
  BODY='{"answers":[{"questionId":'"$Q1_ID"',"userAnswer":["B"]},{"questionId":'"$Q2_ID"',"userAnswer":["A"]}]}'
  post "$USER_API/testpaper/$TESTPAPER_ID/submit" "$DUP2_TOKEN" "$BODY" > "$TMPDIR/dup_submit_$i.json" &
  PIDS+=($!)
done
for pid in "${PIDS[@]}"; do wait $pid; done

SUBMIT_CODES=""
for i in $(seq 1 5); do
  CODE=$(jq -r '.code // "null"' "$TMPDIR/dup_submit_$i.json" 2>/dev/null)
  MSG=$(jq -r '.msg // ""' "$TMPDIR/dup_submit_$i.json" 2>/dev/null)
  SUBMIT_CODES="$SUBMIT_CODES [$CODE:$MSG]"
done
echo "  响应码: $SUBMIT_CODES"

# 检查 DB 中已交卷记录数
SUBMITTED=$(docker exec mysql80 mysql -uroot -proot training -N -e "SELECT COUNT(*) FROM testpaper_records WHERE user_id=$DUP2_ID AND testpaper_id=$TESTPAPER_ID AND submitted_at IS NOT NULL;" 2>/dev/null)
echo "  DB 已交卷记录数: $SUBMITTED (期望 1)"

if [ "$SUBMITTED" = "1" ]; then
  pass "并发交卷未产生重复记录 ($SUBMITTED = 1)"
else
  fail "并发交卷产生重复记录" "DB已交卷记录=$SUBMITTED (期望1)"
fi
echo ""

# ===================================================================
# 场景 7: 并发进度上报压力测试（同一学员高频上报）
# ===================================================================
echo "【场景7: 同一学员高频并发上报（50请求）】"
STRESS_USER="${STUDENT_USERS[3]}"
STRESS_TOKEN="${STUDENT_TOKENS[3]}"
STRESS_ID="${STUDENT_IDS[3]}"

# 清除已有记录
docker exec mysql80 mysql -uroot -proot training -e "DELETE FROM video_records WHERE user_id=$STRESS_ID AND video_id=$VIDEO_ID;" 2>/dev/null
docker exec mysql80 mysql -uroot -proot training -e "DELETE FROM check_logs WHERE user_id=$STRESS_ID AND video_id=$VIDEO_ID;" 2>/dev/null

echo "  学员: $STRESS_USER (id=$STRESS_ID)"
echo "  发送 50 个并发请求 position=600..."

PIDS=()
for i in $(seq 1 50); do
  BODY='{"videoId":'"$VIDEO_ID"',"courseId":'"$COURSE_ID"',"classId":'"$CLASS_ID"',"position":600,"duration":'"$VIDEO_DURATION"',"completed":true}'
  post "$USER_API/progress" "$STRESS_TOKEN" "$BODY" > "$TMPDIR/stress_$i.json" &
  PIDS+=($!)
done
for pid in "${PIDS[@]}"; do wait $pid; done

STRESS_OK=0
for i in $(seq 1 50); do
  CODE=$(jq -r '.code // "null"' "$TMPDIR/stress_$i.json" 2>/dev/null)
  if [ "$CODE" = "0" ]; then STRESS_OK=$((STRESS_OK+1)); fi
done
echo "  成功响应: $STRESS_OK / 50"

# 检查 maxPosition — 50个并发请求不应该让 maxPosition 超过 nextCheckPosition
STRESS_MAX=$(docker exec mysql80 mysql -uroot -proot training -N -e "SELECT max_position FROM video_records WHERE user_id=$STRESS_ID AND video_id=$VIDEO_ID;" 2>/dev/null)
STRESS_COMP=$(docker exec mysql80 mysql -uroot -proot training -N -e "SELECT completed FROM video_records WHERE user_id=$STRESS_ID AND video_id=$VIDEO_ID;" 2>/dev/null)
STRESS_CP=$(docker exec mysql80 mysql -uroot -proot training -N -e "SELECT check_pending FROM video_records WHERE user_id=$STRESS_ID AND video_id=$VIDEO_ID;" 2>/dev/null)
echo "  DB 结果: maxPosition=$STRESS_MAX, completed=$STRESS_COMP, checkPending=$STRESS_CP"

if [ -n "$STRESS_MAX" ] && [ "$STRESS_MAX" -lt 100 ] 2>/dev/null; then
  pass "50并发高压跳跃被限速 (maxPos=$STRESS_MAX < 100)"
else
  fail "50并发高压跳跃未限速" "maxPos=$STRESS_MAX"
fi

if [ "$STRESS_COMP" = "0" ]; then
  pass "50并发高压未标记完成"
else
  fail "50并发高压被标记完成" "completed=$STRESS_COMP"
fi

# 验证 video_records 没有产生重复记录
RECORD_NUM=$(docker exec mysql80 mysql -uroot -proot training -N -e "SELECT COUNT(*) FROM video_records WHERE user_id=$STRESS_ID AND video_id=$VIDEO_ID;" 2>/dev/null)
echo "  DB video_records 记录数: $RECORD_NUM (期望 1)"
if [ "$RECORD_NUM" = "1" ]; then
  pass "高压并发未产生重复 video_records ($RECORD_NUM = 1)"
else
  fail "高压并发产生重复 video_records" "记录数=$RECORD_NUM (期望1)"
fi
echo ""

# ==== 清理 ====
echo "【清理测试数据】"
for i in $(seq 0 $((NUM_STUDENTS-1))); do
  SID="${STUDENT_IDS[$i]}"
  del "$ADMIN_API/user/$SID" "$ADMIN_TOKEN" > /dev/null 2>&1
done
echo "  ✅ 清理 $NUM_STUDENTS 个测试学员"
rm -rf "$TMPDIR"
echo ""

# ==== 汇总 ====
echo "==================================================================="
echo "  并发测试汇总"
echo "==================================================================="
echo "  通过: $PASS"
echo "  失败: $FAIL"
echo "  总计: $((PASS+FAIL))"
echo ""
if [ ${#FAILED_CASES[@]} -gt 0 ]; then
  echo "  失败用例:"
  for c in "${FAILED_CASES[@]}"; do echo "    ❌ $c"; done
  echo ""
fi
if [ $FAIL -eq 0 ]; then echo "  🎉 全部通过！"; else echo "  ⚠️  有 $FAIL 个用例失败"; fi
exit $FAIL
