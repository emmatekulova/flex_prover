/** @type {import('next').NextConfig} */
const nextConfig = {
  typescript: {
    ignoreBuildErrors: true,
  },
  images: {
    unoptimized: true,
  },
  allowedDevOrigins: ['10.105.176.85'],
  turbopack: {
    resolveAlias: {
      'jspdf': 'jspdf/dist/jspdf.es.min.js',
    },
  },
  webpack: (config, { isServer }) => {
    config.externals.push('pino-pretty', 'lokijs', 'encoding')
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
