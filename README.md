# 培训系统（Gin + Vue3）

一套面向企业内部培训的三端系统：Go 后端 + 管理后台 + 学员学习端，覆盖 RBAC 权限、课程/视频管理、班级分配、学习进度追踪与防作弊、统计看板等完整能力。

## 三端结构

| 目录 | 端 | 技术栈 |
|------|----|--------|
| `api/` | Go 后端 | Gin + GORM + MySQL + JWT + 自定义 RBAC + jsoniter |
| `back/` | 管理后台 | Vue3 + Vite + Element Plus + Pinia + TypeScript |
| `front/` | 学员学习/考试端 | Vue3 + Vite + Element Plus + Pinia + TypeScript |

## 默认账号
- 管理员：`admin / admin123`（超管，bypass 权限校验）
- 学员：前台注册或后台批量导入（初始密码 = 手机号）

## 数据初始化
SET FOREIGN_KEY_CHECKS = 0;

TRUNCATE TABLE class_courses;
TRUNCATE TABLE class_users;
TRUNCATE TABLE classes;
TRUNCATE TABLE video_records;
TRUNCATE TABLE videos;
TRUNCATE TABLE courses;

SET FOREIGN_KEY_CHECKS = 1;

## 部署
请参考DEPLOY.md

## 使用Docker
请参考DOCKER_DEPLOY.md
