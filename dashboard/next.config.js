/** @type {import('next').NextConfig} */
const path = require('path');

const nextConfig = {
  outputFileTracingRoot: path.join(__dirname, '../'),
  async rewrites() {
    return [
      {
        source: '/api/v1/:path*',
        destination: 'http://localhost:8080/api/v1/:path*',
      },
      {
        source: '/health',
        destination: 'http://localhost:8080/health',
      },
      {
        source: '/swagger/:path*',
        destination: 'http://localhost:8080/swagger/:path*',
      },
    ];
  },
};

module.exports = nextConfig;