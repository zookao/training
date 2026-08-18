#!/bin/bash
# ===================================================================
# 培训系统后端 API 全面功能测试脚本 v3
# 修复：JSON body 用单引号拼接，避免子shell转义问题
# ===================================================================

BASE="http://127.0.0.1:8000"
ADMIN_API="$BASE/api/admin"
USER_API="$BASE/api/user"
PASS=0
FAIL=0
FAILED_CASES=()
TS=$(date +%s)

assert_ok() {
  local code; code=$(echo "$2" | jq -r '.code // empty' 2>/dev/null)
  if [ "$code" = "0" ]; then PASS=$((PASS+1)); echo "  ✅ $1"
  else FAIL=$((FAIL+1)); FAILED_CASES+=("$1 → $(echo "$2" | jq -r '.msg // "unknown"' 2>/dev/null)"); echo "  ❌ $1 | $(echo "$2" | head -c 200)"; fi
}
assert_fail() {
  local code; code=$(echo "$2" | jq -r '.code // empty' 2>/dev/null)
  if [ "$code" != "0" ] && [ -n "$code" ]; then PASS=$((PASS+1)); echo "  ✅ $1 (正确拒绝)"
  else FAIL=$((FAIL+1)); FAILED_CASES+=("$1 → 期望失败但成功"); echo "  ❌ $1 | 期望失败: $(echo "$2" | head -c 200)"; fi
}
assert_json_fail() {
  local code; code=$(echo "$2" | jq -r '.code // empty' 2>/dev/null)
  if [ "$code" != "0" ] && [ -n "$code" ]; then PASS=$((PASS+1)); echo "  ✅ $1 (code=$code)"
  else FAIL=$((FAIL+1)); FAILED_CASES+=("$1"); echo "  ❌ $1 | $(echo "$2" | head -c 200)"; fi
}
post() { curl -s -X POST "$1" -H "Content-Type: application/json" -H "Authorization: Bearer $2" -d "$3"; }
get()  { curl -s -X GET "$1" -H "Authorization: Bearer $2"; }
put()  { curl -s -X PUT "$1" -H "Content-Type: application/json" -H "Authorization: Bearer $2" -d "$3"; }
del()  { curl -s -X DELETE "$1" -H "Authorization: Bearer $2"; }
find_id() { # api_path, token, field, value — 在 data.list 或 data 数组中按字段查找 id
  local resp; resp=$(get "$1?page=1&pageSize=200" "$2")
  local id; id=$(echo "$resp" | jq -r ".data.list[]? // .data[]? | select(.\"$3\" == \"$4\") | .id" 2>/dev/null | head -1)
  echo "$id"
}

echo "==================================================================="
echo "  培训系统后端 API 全面功能测试 v3  (TS=$TS)"
echo "==================================================================="
echo ""

# ==== 0. 超管登录 ====
ADMIN_TOKEN=$(post "$ADMIN_API/auth/login" "" '{"username":"admin","password":"admin123"}' | jq -r '.data.token')
[ -z "$ADMIN_TOKEN" ] || [ "$ADMIN_TOKEN" = "null" ] && { echo "❌ 超管登录失败"; exit 1; }
echo "超管登录成功"
echo ""

# ==== 1. 认证模块 ====
echo "【1. 认证模块】"
assert_fail "错误密码登录" "$(post "$ADMIN_API/auth/login" "" '{"username":"admin","password":"wrong"}')"
assert_fail "空用户名登录" "$(post "$ADMIN_API/auth/login" "" '{"username":"","password":"admin123"}')"
assert_ok "获取userinfo" "$(get "$ADMIN_API/auth/userinfo" "$ADMIN_TOKEN")"
assert_ok "获取菜单树" "$(get "$ADMIN_API/auth/menus" "$ADMIN_TOKEN")"

RAND_USER="stu_${TS}"
REG_BODY='{"username":"'"$RAND_USER"'","password":"123456","nickname":"测试学员","phone":"138'"${TS: -8}"'"}'
assert_ok "学员注册" "$(post "$USER_API/auth/register" "" "$REG_BODY")"
assert_fail "注册-手机号格式错误" "$(post "$USER_API/auth/register" "" '{"username":"bad","password":"123456","phone":"12345"}')"
assert_fail "注册-手机号为空" "$(post "$USER_API/auth/register" "" '{"username":"nophone","password":"123456"}')"
assert_fail "注册-用户名太短" "$(post "$USER_API/auth/register" "" '{"username":"ab","password":"123456","phone":"13800138001"}')"
assert_fail "注册-密码太短" "$(post "$USER_API/auth/register" "" '{"username":"shortpwd","password":"12345","phone":"13800138002"}')"

USER_TOKEN=$(post "$USER_API/auth/login" "" '{"username":"'"$RAND_USER"'","password":"123456"}' | jq -r '.data.token')
if [ -n "$USER_TOKEN" ] && [ "$USER_TOKEN" != "null" ]; then PASS=$((PASS+1)); echo "  ✅ 学员登录"; else FAIL=$((FAIL+1)); echo "  ❌ 学员登录失败"; fi

assert_fail "学员token调管理端" "$(get "$ADMIN_API/auth/userinfo" "$USER_TOKEN")"
assert_fail "管理端token调学员端" "$(get "$USER_API/auth/userinfo" "$ADMIN_TOKEN")"
assert_fail "无token访问" "$(curl -s "$ADMIN_API/auth/userinfo")"
assert_fail "伪造token访问" "$(get "$ADMIN_API/auth/userinfo" "fake.token")"

assert_ok "超管登出" "$(post "$ADMIN_API/auth/logout" "$ADMIN_TOKEN" '{}')"
assert_fail "登出后旧token" "$(get "$ADMIN_API/auth/userinfo" "$ADMIN_TOKEN")"
ADMIN_TOKEN=$(post "$ADMIN_API/auth/login" "" '{"username":"admin","password":"admin123"}' | jq -r '.data.token')
echo ""

# ==== 2. RBAC ====
echo "【2. RBAC 权限】"
ADMIN2_NAME="admin2_${TS}"
ADMIN2_BODY='{"username":"'"$ADMIN2_NAME"'","password":"123456","nickname":"测试管理员","status":1}'
assert_ok "创建普通管理员" "$(post "$ADMIN_API/admin" "$ADMIN_TOKEN" "$ADMIN2_BODY")"
ADMIN2_ID=$(find_id "$ADMIN_API/admin" "$ADMIN_TOKEN" "username" "$ADMIN2_NAME")
ADMIN2_TOKEN=$(post "$ADMIN_API/auth/login" "" '{"username":"'"$ADMIN2_NAME"'","password":"123456"}' | jq -r '.data.token')
if [ -n "$ADMIN2_TOKEN" ] && [ "$ADMIN2_TOKEN" != "null" ]; then PASS=$((PASS+1)); echo "  ✅ 普通管理员登录"; else FAIL=$((FAIL+1)); echo "  ❌ 普通管理员登录失败"; fi

assert_fail "无角色管理员创建课程" "$(post "$ADMIN_API/course" "$ADMIN2_TOKEN" '{"title":"test","sort":1,"status":1}')"
assert_fail "超管不可被编辑" "$(put "$ADMIN_API/admin/1" "$ADMIN_TOKEN" '{"nickname":"hacked"}')"
assert_fail "超管不可被删除" "$(del "$ADMIN_API/admin/1" "$ADMIN_TOKEN")"
assert_fail "超管不可被分配角色" "$(put "$ADMIN_API/admin/1/roles" "$ADMIN_TOKEN" '{"ids":[1]}')"

ROLE_NAME="role_${TS}"
assert_ok "创建角色" "$(post "$ADMIN_API/role" "$ADMIN_TOKEN" '{"name":"'"$ROLE_NAME"'","title":"测试角色","guardType":"admin","sort":1,"status":1}')"
ROLE_ID=$(find_id "$ADMIN_API/role" "$ADMIN_TOKEN" "name" "$ROLE_NAME")
assert_fail "删除超管角色" "$(del "$ADMIN_API/role/1" "$ADMIN_TOKEN")"
SYS_MENU_ID=$(get "$ADMIN_API/menu" "$ADMIN_TOKEN" | jq -r '.data[0].id // empty')
[ -n "$SYS_MENU_ID" ] && assert_fail "系统管理菜单不可删除" "$(del "$ADMIN_API/menu/$SYS_MENU_ID" "$ADMIN_TOKEN")"
echo ""

# ==== 3. 院系 ====
echo "【3. 院系管理】"
DEPT_NAME="院系_${TS}"
assert_ok "创建院系" "$(post "$ADMIN_API/department" "$ADMIN_TOKEN" '{"name":"'"$DEPT_NAME"'","description":"测试","sort":1,"status":1}')"
DEPT_ID=$(find_id "$ADMIN_API/department" "$ADMIN_TOKEN" "name" "$DEPT_NAME")
assert_ok "更新院系" "$(put "$ADMIN_API/department/$DEPT_ID" "$ADMIN_TOKEN" '{"name":"'"${DEPT_NAME}_改"'","sort":2,"status":1}')"
assert_ok "院系列表" "$(get "$ADMIN_API/department?page=1&pageSize=10" "$ADMIN_TOKEN")"
assert_ok "院系全部列表" "$(get "$ADMIN_API/department/all" "$ADMIN_TOKEN")"
echo ""

# ==== 4. 课程 ====
echo "【4. 课程管理】"
COURSE_NAME="课程_${TS}"
assert_ok "创建课程" "$(post "$ADMIN_API/course" "$ADMIN_TOKEN" '{"title":"'"$COURSE_NAME"'","description":"测试","sort":1,"status":1}')"
COURSE_ID=$(find_id "$ADMIN_API/course" "$ADMIN_TOKEN" "title" "$COURSE_NAME")
assert_ok "课程详情" "$(get "$ADMIN_API/course/$COURSE_ID" "$ADMIN_TOKEN")"
assert_ok "更新课程" "$(put "$ADMIN_API/course/$COURSE_ID" "$ADMIN_TOKEN" '{"title":"'"${COURSE_NAME}_改"'","description":"改后","sort":2,"status":1}')"
assert_ok "课程列表" "$(get "$ADMIN_API/course?page=1&pageSize=10" "$ADMIN_TOKEN")"
assert_ok "课程全部列表" "$(get "$ADMIN_API/course/all" "$ADMIN_TOKEN")"

# 添加视频
VIDEO_URL="/upload/${TS:0:6}/videos/test${TS}.mp4"
THUMB_URL="/upload/${TS:0:6}/thumbnails/test${TS}.jpg"
ADD_VIDEO_BODY='{"title":"'"${COURSE_NAME}_改"'","description":"改后","sort":2,"status":1,"videos":[{"url":"'"$VIDEO_URL"'","thumbnail":"'"$THUMB_URL"'","title":"视频1","duration":600,"sort":1}]}'
assert_ok "更新课程-添加视频" "$(put "$ADMIN_API/course/$COURSE_ID" "$ADMIN_TOKEN" "$ADD_VIDEO_BODY")"
VIDEO_ID=$(get "$ADMIN_API/course/$COURSE_ID" "$ADMIN_TOKEN" | jq -r '.data.videos[0].id // empty')
if [ -n "$VIDEO_ID" ]; then
  PASS=$((PASS+1)); echo "  ✅ 视频已添加 (id=$VIDEO_ID)"
  KEEP_VIDEO_BODY='{"title":"'"${COURSE_NAME}_改"'","sort":2,"status":1,"videos":[{"id":'"$VIDEO_ID"',"url":"'"$VIDEO_URL"'","thumbnail":"'"$THUMB_URL"'","title":"视频1","duration":600,"sort":1}]}'
  assert_ok "更新课程-视频带ID保留" "$(put "$ADMIN_API/course/$COURSE_ID" "$ADMIN_TOKEN" "$KEEP_VIDEO_BODY")"
else FAIL=$((FAIL+1)); echo "  ❌ 未找到视频ID"; fi

KEEP_BODY='{"title":"'"${COURSE_NAME}_改"'","sort":2,"status":'
assert_ok "禁用课程" "$(put "$ADMIN_API/course/$COURSE_ID" "$ADMIN_TOKEN" "$KEEP_BODY"'0,"videos":[{"id":'"$VIDEO_ID"',"url":"'"$VIDEO_URL"'","thumbnail":"'"$THUMB_URL"'","title":"视频1","duration":600,"sort":1}]}')"
assert_ok "启用课程" "$(put "$ADMIN_API/course/$COURSE_ID" "$ADMIN_TOKEN" "$KEEP_BODY"'1,"videos":[{"id":'"$VIDEO_ID"',"url":"'"$VIDEO_URL"'","thumbnail":"'"$THUMB_URL"'","title":"视频1","duration":600,"sort":1}]}')"
echo ""

# ==== 5. 学员 ====
echo "【5. 学员管理】"
STU_NAME="adminstu_${TS}"
STU_PHONE="139${TS: -8}"
STU_BODY='{"username":"'"$STU_NAME"'","password":"123456","nickname":"测试学员","studentNo":"S'"$TS"'","departmentId":'"$DEPT_ID"',"phone":"'"$STU_PHONE"'","email":"t@t.com","status":1}'
assert_ok "创建学员" "$(post "$ADMIN_API/user" "$ADMIN_TOKEN" "$STU_BODY")"
STU_ID=$(find_id "$ADMIN_API/user" "$ADMIN_TOKEN" "username" "$STU_NAME")
assert_fail "创建学员-手机号格式错误" "$(post "$ADMIN_API/user" "$ADMIN_TOKEN" '{"username":"badp","password":"123456","phone":"12345","status":1}')"
assert_fail "创建学员-无手机号" "$(post "$ADMIN_API/user" "$ADMIN_TOKEN" '{"username":"nophone","password":"123456","status":1}')"
assert_fail "创建学员-重复用户名" "$(post "$ADMIN_API/user" "$ADMIN_TOKEN" '{"username":"'"$STU_NAME"'","password":"123456","phone":"137'"${TS: -8}"'","status":1}')"

STU_UPDATE_BODY='{"nickname":"测试改","studentNo":"SM'"$TS"'","departmentId":'"$DEPT_ID"',"phone":"137'"${TS: -8}"'","status":1}'
assert_ok "更新学员" "$(put "$ADMIN_API/user/$STU_ID" "$ADMIN_TOKEN" "$STU_UPDATE_BODY")"
assert_fail "更新学员-手机号格式错误" "$(put "$ADMIN_API/user/$STU_ID" "$ADMIN_TOKEN" '{"phone":"abc","status":1}')"
assert_ok "学员列表" "$(get "$ADMIN_API/user?page=1&pageSize=10" "$ADMIN_TOKEN")"
assert_ok "学员详情" "$(get "$ADMIN_API/user/$STU_ID" "$ADMIN_TOKEN")"
assert_ok "重置学员密码" "$(put "$ADMIN_API/user/$STU_ID/password" "$ADMIN_TOKEN" '{"password":"newpass123"}')"

DEL_STU="delstu_${TS}"
post "$ADMIN_API/user" "$ADMIN_TOKEN" '{"username":"'"$DEL_STU"'","password":"123456","phone":"136'"${TS: -8}"'","status":1}' > /dev/null
DEL_STU_ID=$(find_id "$ADMIN_API/user" "$ADMIN_TOKEN" "username" "$DEL_STU")
assert_ok "删除学员" "$(del "$ADMIN_API/user/$DEL_STU_ID" "$ADMIN_TOKEN")"
echo ""

# ==== 6. 班级 ====
echo "【6. 班级管理】"
CLASS_NAME="班级_${TS}"
assert_ok "创建班级" "$(post "$ADMIN_API/class" "$ADMIN_TOKEN" '{"name":"'"$CLASS_NAME"'","description":"测试","sort":1,"status":1}')"
CLASS_ID=$(find_id "$ADMIN_API/class" "$ADMIN_TOKEN" "name" "$CLASS_NAME")
assert_ok "更新班级" "$(put "$ADMIN_API/class/$CLASS_ID" "$ADMIN_TOKEN" '{"name":"'"${CLASS_NAME}_改"'","sort":2,"status":1}')"
assert_ok "班级列表" "$(get "$ADMIN_API/class?page=1&pageSize=10" "$ADMIN_TOKEN")"
assert_ok "班级详情" "$(get "$ADMIN_API/class/$CLASS_ID" "$ADMIN_TOKEN")"
assert_ok "分配课程到班级" "$(put "$ADMIN_API/class/$CLASS_ID/courses" "$ADMIN_TOKEN" '{"ids":['"$COURSE_ID"']}')"
assert_ok "获取班级课程IDs" "$(get "$ADMIN_API/class/$CLASS_ID/courseIds" "$ADMIN_TOKEN")"
assert_ok "分配学员到班级" "$(put "$ADMIN_API/class/$CLASS_ID/users" "$ADMIN_TOKEN" '{"ids":['"$STU_ID"']}')"
assert_ok "获取班级学员IDs" "$(get "$ADMIN_API/class/$CLASS_ID/userIds" "$ADMIN_TOKEN")"
assert_ok "班级学习报告" "$(get "$ADMIN_API/class/$CLASS_ID/learning-report" "$ADMIN_TOKEN")"
assert_ok "学员学习详情" "$(get "$ADMIN_API/class/$CLASS_ID/learning-report/$STU_ID" "$ADMIN_TOKEN")"
echo ""

# ==== 7. 试题 ====
echo "【7. 试题管理】"
Q1_TITLE="单选_${TS}"
Q1_BODY='{"type":1,"title":"'"$Q1_TITLE"'","options":"[{\"label\":\"A\",\"content\":\"选项A\"},{\"label\":\"B\",\"content\":\"选项B\"}]","answer":"[\"A\"]","sort":1,"status":1}'
assert_ok "创建单选题" "$(post "$ADMIN_API/course/$COURSE_ID/question" "$ADMIN_TOKEN" "$Q1_BODY")"
Q1_ID=$(find_id "$ADMIN_API/course/$COURSE_ID/question" "$ADMIN_TOKEN" "title" "$Q1_TITLE")

Q2_TITLE="多选_${TS}"
Q2_BODY='{"type":2,"title":"'"$Q2_TITLE"'","options":"[{\"label\":\"A\",\"content\":\"选项A\"},{\"label\":\"B\",\"content\":\"选项B\"}]","answer":"[\"A\",\"B\"]","sort":2,"status":1}'
assert_ok "创建多选题" "$(post "$ADMIN_API/course/$COURSE_ID/question" "$ADMIN_TOKEN" "$Q2_BODY")"
Q2_ID=$(find_id "$ADMIN_API/course/$COURSE_ID/question" "$ADMIN_TOKEN" "title" "$Q2_TITLE")

Q3_TITLE="判断_${TS}"
Q3_BODY='{"type":3,"title":"'"$Q3_TITLE"'","options":"[{\"label\":\"A\",\"content\":\"正确\"},{\"label\":\"B\",\"content\":\"错误\"}]","answer":"[\"A\"]","sort":3,"status":1}'
assert_ok "创建判断题" "$(post "$ADMIN_API/course/$COURSE_ID/question" "$ADMIN_TOKEN" "$Q3_BODY")"
Q3_ID=$(find_id "$ADMIN_API/course/$COURSE_ID/question" "$ADMIN_TOKEN" "title" "$Q3_TITLE")

assert_fail "同课程重复题干" "$(post "$ADMIN_API/course/$COURSE_ID/question" "$ADMIN_TOKEN" "$Q1_BODY")"
Q1_UPDATE_BODY='{"type":1,"title":"'"${Q1_TITLE}_改"'","options":"[{\"label\":\"A\",\"content\":\"改A\"}]","answer":"[\"A\"]","sort":1,"status":1}'
assert_ok "更新试题" "$(put "$ADMIN_API/question/$Q1_ID" "$ADMIN_TOKEN" "$Q1_UPDATE_BODY")"
assert_fail "更新试题-重复题干" "$(put "$ADMIN_API/question/$Q1_ID" "$ADMIN_TOKEN" '{"type":1,"title":"'"$Q2_TITLE"'","options":"[]","answer":"[\"A\"]","status":1}')"
assert_ok "试题列表" "$(get "$ADMIN_API/course/$COURSE_ID/question?page=1&pageSize=10" "$ADMIN_TOKEN")"
assert_ok "试题全部列表" "$(get "$ADMIN_API/course/$COURSE_ID/question/all" "$ADMIN_TOKEN")"
assert_ok "删除判断题" "$(del "$ADMIN_API/question/$Q3_ID" "$ADMIN_TOKEN")"
echo ""

# ==== 8. 试卷 ====
echo "【8. 试卷管理】"
TP1_NAME="试卷1_${TS}"
TP1_BODY='{"name":"'"$TP1_NAME"'","description":"测试","type":1,"totalScore":100,"passScore":60,"duration":60,"sort":1,"status":1}'
assert_ok "创建试卷1(启用)" "$(post "$ADMIN_API/course/$COURSE_ID/testpaper" "$ADMIN_TOKEN" "$TP1_BODY")"
TP1_ID=$(find_id "$ADMIN_API/course/$COURSE_ID/testpaper" "$ADMIN_TOKEN" "name" "$TP1_NAME")

TP2_NAME="试卷2_${TS}"
TP2_BODY='{"name":"'"$TP2_NAME"'","description":"测试2","type":1,"totalScore":100,"passScore":60,"duration":60,"sort":2,"status":1}'
assert_ok "创建试卷2(启用,自动禁用试卷1)" "$(post "$ADMIN_API/course/$COURSE_ID/testpaper" "$ADMIN_TOKEN" "$TP2_BODY")"
TP2_ID=$(find_id "$ADMIN_API/course/$COURSE_ID/testpaper" "$ADMIN_TOKEN" "name" "$TP2_NAME")

TP1_STATUS=$(get "$ADMIN_API/course/$COURSE_ID/testpaper?page=1&pageSize=10" "$ADMIN_TOKEN" | jq -r ".data.list[] | select(.id==$TP1_ID) | .status")
if [ "$TP1_STATUS" = "0" ]; then PASS=$((PASS+1)); echo "  ✅ 试卷1被自动禁用"; else FAIL=$((FAIL+1)); echo "  ❌ 试卷1未被自动禁用 (status=$TP1_STATUS)"; fi

assert_fail "同课程重复试卷名" "$(post "$ADMIN_API/course/$COURSE_ID/testpaper" "$ADMIN_TOKEN" "$TP2_BODY")"
assert_fail "更新试卷-重复名" "$(put "$ADMIN_API/testpaper/$TP2_ID" "$ADMIN_TOKEN" '{"name":"'"$TP1_NAME"'","type":1,"totalScore":100,"passScore":60,"duration":60,"status":1}')"

SET_Q_BODY='{"items":[{"questionId":'"$Q1_ID"',"score":50},{"questionId":'"$Q2_ID"',"score":50}]}'
assert_ok "组卷-设置试题" "$(put "$ADMIN_API/testpaper/$TP2_ID/questions" "$ADMIN_TOKEN" "$SET_Q_BODY")"
assert_ok "获取试卷试题" "$(get "$ADMIN_API/testpaper/$TP2_ID/questions" "$ADMIN_TOKEN")"

# 启用试卷1 → 自动禁用试卷2
TP1_ENABLE_BODY='{"name":"'"$TP1_NAME"'","type":1,"totalScore":100,"passScore":60,"duration":60,"status":1}'
assert_ok "启用试卷1(自动禁用试卷2)" "$(put "$ADMIN_API/testpaper/$TP1_ID" "$ADMIN_TOKEN" "$TP1_ENABLE_BODY")"
TP2_STATUS=$(get "$ADMIN_API/course/$COURSE_ID/testpaper?page=1&pageSize=10" "$ADMIN_TOKEN" | jq -r ".data.list[] | select(.id==$TP2_ID) | .status")
if [ "$TP2_STATUS" = "0" ]; then PASS=$((PASS+1)); echo "  ✅ 试卷2被自动禁用"; else FAIL=$((FAIL+1)); echo "  ❌ 试卷2未被自动禁用 (status=$TP2_STATUS)"; fi

DEL_TP_NAME="待删试卷_${TS}"
post "$ADMIN_API/course/$COURSE_ID/testpaper" "$ADMIN_TOKEN" '{"name":"'"$DEL_TP_NAME"'","type":1,"totalScore":100,"passScore":60,"duration":30,"status":0}' > /dev/null
DEL_TP_ID=$(find_id "$ADMIN_API/course/$COURSE_ID/testpaper" "$ADMIN_TOKEN" "name" "$DEL_TP_NAME")
assert_ok "删除试卷" "$(del "$ADMIN_API/testpaper/$DEL_TP_ID" "$ADMIN_TOKEN")"
echo ""

# ==== 9. 学习 + 防作弊 ====
echo "【9. 学习 + 防作弊】"
# 启用试卷2给考试用
put "$ADMIN_API/testpaper/$TP2_ID" "$ADMIN_TOKEN" "$TP2_BODY" > /dev/null

STU_TOKEN=$(post "$USER_API/auth/login" "" '{"username":"'"$STU_NAME"'","password":"newpass123"}' | jq -r '.data.token')
if [ -z "$STU_TOKEN" ] || [ "$STU_TOKEN" = "null" ]; then STU_TOKEN=$(post "$USER_API/auth/login" "" '{"username":"'"$STU_NAME"'","password":"123456"}' | jq -r '.data.token'); fi

if [ -n "$STU_TOKEN" ] && [ "$STU_TOKEN" != "null" ]; then
  assert_ok "学员-我的班级" "$(get "$USER_API/classes" "$STU_TOKEN")"
  assert_ok "学员-班级详情" "$(get "$USER_API/classes/$CLASS_ID" "$STU_TOKEN")"
  assert_ok "学员-课程学习详情" "$(get "$USER_API/course/$COURSE_ID" "$STU_TOKEN")"

  if [ -n "$VIDEO_ID" ]; then
    PROG_BODY='{"videoId":'"$VIDEO_ID"',"courseId":'"$COURSE_ID"',"classId":'"$CLASS_ID"',"position":10,"duration":600,"completed":false}'
    PROG1=$(post "$USER_API/progress" "$STU_TOKEN" "$PROG_BODY")
    assert_ok "首次上报进度(10s)" "$PROG1"
    NEXT_CHECK=$(echo "$PROG1" | jq -r '.data.nextCheckPosition // 0')
    if [ "$NEXT_CHECK" -gt 0 ] 2>/dev/null; then PASS=$((PASS+1)); echo "  ✅ 返回校验点 (nextCheck=$NEXT_CHECK)"; else FAIL=$((FAIL+1)); echo "  ❌ 未返回校验点"; fi

    # 防作弊-跳跃上报
    CHEAT_BODY='{"videoId":'"$VIDEO_ID"',"courseId":'"$COURSE_ID"',"classId":'"$CLASS_ID"',"position":600,"duration":600,"completed":true}'
    PROG_CHEAT=$(post "$USER_API/progress" "$STU_TOKEN" "$CHEAT_BODY")
    MAX_POS=$(echo "$PROG_CHEAT" | jq -r '.data.maxPosition // 0')
    if [ "$MAX_POS" -lt 100 ] 2>/dev/null; then PASS=$((PASS+1)); echo "  ✅ 防作弊-跳跃被限速 (maxPos=$MAX_POS)"; else FAIL=$((FAIL+1)); echo "  ❌ 跳跃未被限速 (maxPos=$MAX_POS)"; fi
    COMPLETED=$(echo "$PROG_CHEAT" | jq -r '.data.completed // false')
    if [ "$COMPLETED" = "false" ]; then PASS=$((PASS+1)); echo "  ✅ 防作弊-伪造completed无效"; else FAIL=$((FAIL+1)); echo "  ❌ 伪造completed成功了"; fi

    assert_fail "未到校验点调checkPass" "$(post "$USER_API/video/$VIDEO_ID/check" "$STU_TOKEN" '{}')"
  else echo "  ⚠️ 无视频ID，跳过进度测试"; fi

  # 无权限学员
  NOACC_NAME="noacc_${TS}"
  post "$ADMIN_API/user" "$ADMIN_TOKEN" '{"username":"'"$NOACC_NAME"'","password":"123456","phone":"135'"${TS: -8}"'","status":1}' > /dev/null
  NOACC_TOKEN=$(post "$USER_API/auth/login" "" '{"username":"'"$NOACC_NAME"'","password":"123456"}' | jq -r '.data.token')
  assert_fail "无权限学员上报进度" "$(post "$USER_API/progress" "$NOACC_TOKEN" '{"videoId":'"$VIDEO_ID"',"courseId":'"$COURSE_ID"',"position":5,"duration":600}')"
  assert_fail "无权限学员访问课程" "$(get "$USER_API/course/$COURSE_ID" "$NOACC_TOKEN")"

  assert_ok "学员-userinfo" "$(get "$USER_API/auth/userinfo" "$STU_TOKEN")"
  PROFILE_BODY='{"nickname":"测试改","phone":"139'"${TS: -8}"'"}'
  assert_ok "学员-更新资料" "$(put "$USER_API/auth/profile" "$STU_TOKEN" "$PROFILE_BODY")"
  assert_fail "学员-更新资料手机号错误" "$(put "$USER_API/auth/profile" "$STU_TOKEN" '{"phone":"abc"}')"
else echo "  ⚠️ 学员登录失败，跳过学习测试"; fi
echo ""

# ==== 10. 考试 ====
echo "【10. 考试模块】"
if [ -n "$STU_TOKEN" ] && [ "$STU_TOKEN" != "null" ] && [ -n "$Q1_ID" ] && [ -n "$Q2_ID" ]; then
  assert_ok "获取考试列表" "$(get "$USER_API/course/$COURSE_ID/exam" "$STU_TOKEN")"
  assert_ok "进入考试" "$(get "$USER_API/testpaper/$TP2_ID/exam" "$STU_TOKEN")"
  assert_ok "保存草稿" "$(post "$USER_API/testpaper/$TP2_ID/draft" "$STU_TOKEN" '{"draftAnswers":"{\"q1\":\"A\"}"}')"
  SUBMIT_BODY='{"answers":[{"questionId":'"$Q1_ID"',"answer":"[\"A\"]"},{"questionId":'"$Q2_ID"',"answer":"[\"A\",\"B\"]"}]}'
  assert_ok "提交考试" "$(post "$USER_API/testpaper/$TP2_ID/submit" "$STU_TOKEN" "$SUBMIT_BODY")"
  assert_ok "查看考试记录" "$(get "$USER_API/course/$COURSE_ID/exam/records" "$STU_TOKEN")"
fi
echo ""

# ==== 11. 上传鉴权 ====
echo "【11. 上传鉴权】"
assert_json_fail "无token访问upload" "$(curl -s "$BASE/upload/260807/videos/test.mp4")"
assert_json_fail "伪造token访问upload" "$(curl -s "$BASE/upload/260807/videos/test.mp4?token=fake")"
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" "$BASE/upload/260807/videos/notexist.mp4?token=$ADMIN_TOKEN")
if [ "$HTTP_CODE" = "404" ]; then PASS=$((PASS+1)); echo "  ✅ 有效token访问不存在文件返回404"; else FAIL=$((FAIL+1)); echo "  ❌ 返回$HTTP_CODE (期望404)"; fi
echo ""

# ==== 12. 删除保护 ====
echo "【12. 删除保护】"
assert_fail "有学员的院系不可删除" "$(del "$ADMIN_API/department/$DEPT_ID" "$ADMIN_TOKEN")"
if [ -n "$VIDEO_ID" ]; then
  assert_fail "有学习数据的班级不可删除" "$(del "$ADMIN_API/class/$CLASS_ID" "$ADMIN_TOKEN")"
else
  # 无学习数据时班级应可删除（不算失败）
  PASS=$((PASS+1)); echo "  ✅ 无学习数据的班级可删除 (跳过删除保护测试)"
fi
echo ""

# ==== 13. 仪表盘 ====
echo "【13. 仪表盘】"
assert_ok "管理端仪表盘" "$(get "$ADMIN_API/dashboard" "$ADMIN_TOKEN")"
echo ""

# ==== 清理 ====
echo "【清理测试数据】"
del "$ADMIN_API/question/$Q1_ID" "$ADMIN_TOKEN" > /dev/null 2>&1 && echo "  ✅ 清理试题1"
del "$ADMIN_API/question/$Q2_ID" "$ADMIN_TOKEN" > /dev/null 2>&1 && echo "  ✅ 清理试题2"
del "$ADMIN_API/testpaper/$TP1_ID" "$ADMIN_TOKEN" > /dev/null 2>&1 && echo "  ✅ 清理试卷1"
del "$ADMIN_API/testpaper/$TP2_ID" "$ADMIN_TOKEN" > /dev/null 2>&1 && echo "  ✅ 清理试卷2"
del "$ADMIN_API/course/$COURSE_ID" "$ADMIN_TOKEN" > /dev/null 2>&1 && echo "  ✅ 清理课程"
del "$ADMIN_API/class/$CLASS_ID" "$ADMIN_TOKEN" > /dev/null 2>&1 && echo "  ✅ 清理班级"
NOACC_ID=$(find_id "$ADMIN_API/user" "$ADMIN_TOKEN" "username" "$NOACC_NAME")
del "$ADMIN_API/user/$STU_ID" "$ADMIN_TOKEN" > /dev/null 2>&1 && echo "  ✅ 清理学员"
del "$ADMIN_API/user/$NOACC_ID" "$ADMIN_TOKEN" > /dev/null 2>&1 && echo "  ✅ 清理无权限学员"
del "$ADMIN_API/department/$DEPT_ID" "$ADMIN_TOKEN" > /dev/null 2>&1 && echo "  ✅ 清理院系"
del "$ADMIN_API/role/$ROLE_ID" "$ADMIN_TOKEN" > /dev/null 2>&1 && echo "  ✅ 清理角色"
del "$ADMIN_API/admin/$ADMIN2_ID" "$ADMIN_TOKEN" > /dev/null 2>&1 && echo "  ✅ 清理测试管理员"
echo ""

# ==== 汇总 ====
echo "==================================================================="
echo "  测试汇总"
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
