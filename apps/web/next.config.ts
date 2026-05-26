import type { NextConfig } from 'next';

const nextConfig: NextConfig = {
  transpilePackages: ['@daily-it-podcast/core', '@daily-it-podcast/drive'],
};

export default nextConfig;
