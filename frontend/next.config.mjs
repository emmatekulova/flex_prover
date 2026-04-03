/** @type {import('next').NextConfig} */
const nextConfig = {
  typescript: {
    ignoreBuildErrors: true,
  },
  images: {
    unoptimized: true,
  },
  //allowedDevOrigins: ['xxx.xxx.x.xx'],
}

export default nextConfig
