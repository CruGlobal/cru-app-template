// Bundles src/index.ts to dist/handler.js as CommonJS, which is what the AWS
// Lambda Node runtime and DataDog's wrapper load.
//
// GOTCHA: this package.json says "type": "module", so *locally* node reads a
// bare dist/handler.js as ESM and `require()`ing it fails with "handler is not
// a function". The Lambda image never hits this — its Dockerfile copies
// dist/ into $LAMBDA_TASK_ROOT with no "type": "module" alongside it, so .js
// is CJS there. To invoke the bundle on your machine, copy it to a .cjs first:
//
//   npm run build && cp dist/handler.js dist/handler.cjs \
//     && node -e 'require("./dist/handler.cjs").handler({}).then(console.log)'
import * as esbuild from 'esbuild'

await esbuild.build({
    entryPoints: ['./src/index.ts'],
    bundle: true,
    platform: 'node',
    target: 'node24',
    outfile: 'dist/handler.js',
    sourcemap: true,
    format: 'cjs',
})

console.log('Built handler.js')
