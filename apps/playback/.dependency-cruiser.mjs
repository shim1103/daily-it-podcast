/** @type {import('dependency-cruiser').IConfiguration} */
export default {
  forbidden: [
    {
      name: "worker-entities-no-outer",
      severity: "error",
      comment: "Entities は内側のみ",
      from: { path: "^worker/src/entities/" },
      to: {
        path: "^worker/src/(application|infrastructure|controllers|routes|composition)/|^contracts/",
      },
    },
    {
      name: "worker-application-no-outer",
      severity: "error",
      comment: "Application は entities/application/contracts のみ（generator 寄せ）",
      from: { path: "^worker/src/application/" },
      to: {
        path: "^worker/src/(infrastructure|controllers|routes|composition)/",
      },
    },
    {
      name: "worker-infra-ports-only",
      severity: "error",
      comment: "Infrastructure が Application から import してよいのは ports のみ",
      from: { path: "^worker/src/infrastructure/" },
      to: { path: "^worker/src/application/(?!ports/)" },
    },
    {
      name: "worker-infra-no-http-contracts",
      severity: "error",
      comment: "Infrastructure は apps/playback/contracts（HTTP）禁止。Drive は repo 根 contracts",
      from: { path: "^worker/src/infrastructure/" },
      to: {
        path: "^worker/src/(controllers|routes|composition)/|^contracts/",
      },
    },
    {
      name: "worker-controllers-no-outer",
      severity: "error",
      comment: "Controller は entities/application/contracts 可",
      from: { path: "^worker/src/controllers/" },
      to: {
        path: "^worker/src/(infrastructure|composition|routes)/",
      },
    },
    {
      name: "worker-routes-only-composition-contracts",
      severity: "error",
      comment: "Route は composition + contracts のみ",
      from: { path: "^worker/src/routes/" },
      to: {
        path: "^worker/src/(entities|application|infrastructure|controllers)/",
      },
    },
    {
      name: "web-utils-no-internal",
      severity: "error",
      comment: "純粋関数層は内部層ゼロ",
      from: { path: "^web/src/utils/" },
      to: {
        path: "^web/src/(api|components|lib|pages|view-models)/|^contracts/",
      },
    },
    {
      name: "web-lib-no-ui-internal",
      severity: "error",
      comment: "External Dependencies は UI/api/utils/view-models 禁止",
      from: { path: "^web/src/lib/" },
      to: {
        path: "^web/src/(api|components|pages|utils|view-models)/",
      },
    },
    {
      name: "web-api-no-ui",
      severity: "error",
      comment: "API Client は contracts + utils 可",
      from: { path: "^web/src/api/" },
      to: {
        path: "^web/src/(components|view-models|lib|pages)/",
      },
    },
    {
      name: "web-viewmodels-no-components-contracts",
      severity: "error",
      comment: "ViewModel は api/utils/lib 可。components/contracts 禁止",
      from: { path: "^web/src/view-models/" },
      to: {
        path: "^web/src/(components|pages)/|^contracts/",
      },
    },
    {
      name: "web-feature-no-api-lib-contracts",
      severity: "error",
      comment: "Feature Component は view-models/utils/feature/primitive 可",
      from: { path: "^web/src/components/feature/" },
      to: {
        path: "^web/src/(api|lib|pages)/|^contracts/",
      },
    },
    {
      name: "web-primitive-strict",
      severity: "error",
      comment: "Primitive は utils + 他 Primitive のみ",
      from: { path: "^web/src/components/primitive/" },
      to: {
        path: "^web/src/(api|lib|pages|view-models|components/feature)/|^contracts/",
      },
    },
    {
      name: "cli-no-other-runtime",
      severity: "error",
      comment:
        "cli（Node の運用 CLI）は web・worker・contracts を import しない（runtime は互いに import しない）。例外は出力の実体 1 点（worker/src/routes/logger.ts）だけ",
      from: { path: "^cli/" },
      to: {
        path: "^(web|worker)/src/|^contracts/",
        pathNot: "^worker/src/routes/logger\\.ts$",
      },
    },
    {
      name: "cli-only-main-uses-output",
      severity: "error",
      comment: "出力の実体（logger.ts）を呼んでよいのは結線の main.ts だけ。command と storage-state は注入された関数を呼ぶ",
      from: { path: "^cli/", pathNot: "^cli/access-session/main\\.ts$" },
      to: { path: "^worker/src/routes/logger\\.ts$" },
    },
    {
      name: "runtime-no-cli",
      severity: "error",
      comment: "web・worker・contracts は cli を import しない",
      from: { path: "^(web|worker)/src/|^contracts/" },
      to: { path: "^cli/" },
    },
    {
      name: "cli-storage-state-pure",
      severity: "error",
      comment: "storage-state は IO を持たない。出力・入口へ依存しない",
      from: { path: "^cli/access-session/storage-state\\.ts$" },
      to: { path: "^cli/access-session/(command|main)\\.ts$" },
    },
  ],
  options: {
    doNotFollow: { path: "node_modules" },
    exclude: {
      path: "\\.(sociable_unit|broad_integration|narrow_integration|contract|system_e2e)\\.test\\.ts$|\\.test\\.ts$",
    },
    tsConfig: { fileName: "tsconfig.json" },
    enhancedResolveOptions: {
      extensions: [".ts", ".tsx", ".js", ".mjs"],
    },
  },
};
