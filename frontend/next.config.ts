import type { NextConfig } from "next";

// When the frontend and Go API run in the same container (single-app
// deployment, e.g. Render), the browser talks only to this origin and the
// Next server proxies API traffic to the backend process. The frontend must
// then be built with NEXT_PUBLIC_API_URL="" so the client uses relative URLs.
const apiProxyTarget =
  process.env.API_PROXY_TARGET ?? "http://127.0.0.1:8090";

const nextConfig: NextConfig = {
  output: "standalone",
  async rewrites() {
    return [
      { source: "/api/:path*", destination: `${apiProxyTarget}/api/:path*` },
      { source: "/healthz", destination: `${apiProxyTarget}/healthz` },
    ];
  },
};

export default nextConfig;
