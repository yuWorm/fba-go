# Agent Instructions

- 新增或修改代码时，必要注释不能省略，必须为复杂逻辑、重要约束、边界条件和非显而易见的行为添加说明；避免添加只复述代码含义的空泛注释。
- 原 Python 项目参考源码路径：`sources/fastapi-best-architecture/`；进行接口、模型、路由、业务行为迁移或对齐时，优先参考该目录下的实现。
- 涉及 FBA Go 框架、插件开发、模板生成或 admin template 维护时，优先使用 `templates/fba-go-template/skills/` 下对应 skill 及其 `references/`，保持 AI 工程化规则和实现一致。
