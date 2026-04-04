# flex_prover

## Quick start

1. Go to frontend:
	- `cd frontend`

2. Install dependencies:
	- `npm install`

3. Run development server:
	- `npm run dev`

## LAN dev access (optional)

If you test from another device on your local network, add your local IP to `allowedDevOrigins` in [frontend/next.config.mjs](frontend/next.config.mjs):

- `allowedDevOrigins: ['192.168.x.x']`