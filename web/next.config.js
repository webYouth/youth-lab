// Next.js runtime config for /blog deployment path.
/** @type {import('next').NextConfig} */
const nextConfig = {
  basePath: "/blog",
  trailingSlash: true,
  output: "standalone"
};

module.exports = nextConfig;
