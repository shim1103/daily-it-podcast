/** @type {import('dependency-cruiser').IConfiguration} */
export default {
  forbidden: [
    {
      name: "worker-entities-no-outer",
      severity: "error",
      comment:
        "Entities は内側のみ。外部依存ゼロの最内側 domain の語彙は型に使ってよい（逆は domain-no-outer が禁じる）",
      from: { path: "^worker/src/entities/" },
      to: {
        path: "^worker/src/(application|infrastructure|controllers|routes|composition)/|^contracts/",
      },
    },
    {
      name: "domain-no-outer",
      severity: "error",
      comment:
        "domain は web・worker・contracts を import しない。web と worker が共有する、http 非依存の語彙",
      from: { path: "^domain/" },
      to: { path: "^(worker|web|contracts)/" },
    },
    {
      name: "domain-no-npm",
      severity: "error",
      comment: "domain は外部依存ゼロ（zod など npm の package を import しない）",
      from: { path: "^domain/" },
      to: {
        dependencyTypes: [
          "npm",
          "npm-dev",
          "npm-optional",
          "npm-peer",
          "npm-bundled",
          "npm-no-pkg",
          "npm-unknown",
        ],
      },
    },
    {
      name: "contracts-only-domain",
      severity: "error",
      comment: "contracts（HTTP の wire 形）が import してよいのは domain のみ。worker・web は不可",
      from: { path: "^contracts/" },
      to: { path: "^(worker|web)/" },
    },
    {
      name: "worker-application-no-outer",
      severity: "error",
      comment: "Application は entities/application のみ",
      from: { path: "^worker/src/application/" },
      to: {
        path: "^worker/src/(infrastructure|controllers|routes|composition)/",
      },
    },
    {
      name: "worker-application-no-http-contracts",
      severity: "error",
      comment:
        "Application は境界共有型（apps/playback/contracts）を import しない。UseCase の入出力は application の型で受け、Controller が契約の型へ写す。repo 根の contracts は別物で対象外",
      from: { path: "^worker/src/application/" },
      to: { path: "^contracts/" },
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
      comment: "Primitive は utils + 他 Primitive のみ（domain knowledge zero なので domain も不可）",
      from: { path: "^web/src/components/primitive/" },
      to: {
        path: "^web/src/(api|lib|pages|view-models|components/feature)/|^contracts/|^domain/",
      },
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
