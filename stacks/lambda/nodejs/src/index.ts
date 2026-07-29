// Minimal Cru Lambda handler — invoked by an event (API Gateway, SQS, an
// EventBridge schedule, etc.), not a web server with a port. esbuild bundles
// this to dist/handler.js, and the container's CMD points at the DataDog
// wrapper which in turn calls this `handler` export (see Dockerfile).
//
// Grow it into whatever the app needs: read `event`, do work, return a result.
// Keep the `handler` signature and a fast, successful return — that's what the
// platform invokes and judges health by.
export const handler = async (event: any): Promise<{ statusCode: number; body: string }> => {
  return {
    statusCode: 200,
    body: JSON.stringify({ status: "ok" }),
  };
};
