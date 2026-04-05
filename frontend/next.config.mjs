/** @type {import('next').NextConfig} */
const nextConfig = {
  typescript: {
    ignoreBuildErrors: true,
  },
  images: {
    unoptimized: true,
  },
  allowedDevOrigins: ['10.105.176.70'],
  turbopack: {
    resolveAlias: {
      'jspdf': 'jspdf/dist/jspdf.es.min.js',
    },
  },
  webpack: (config, { isServer }) => {
    config.externals.push('pino-pretty', 'lokijs', 'encoding')
    if (!isServer) {
      config.resolve.fallback = {
        ...config.resolve.fallback,
        fs: false, net: false, tls: false, dns: false,
      }
    }
    if (isServer) {
      config.externals.push('jspdf', 'html2canvas')
    } else {
      config.resolve.alias = {
        ...config.resolve.alias,
        'jspdf': 'jspdf/dist/jspdf.es.min.js',
      }
    }
    return config
  },
}

export default nextConfig
