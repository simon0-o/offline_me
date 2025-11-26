# Work Time Tracker - Frontend

This is the Next.js frontend for the Work Time Tracker application.

## Tech Stack

- **Next.js 16** - React framework with static export
- **TypeScript** - Type-safe development
- **Tailwind CSS** - Utility-first CSS framework
- **React 19** - Latest React features

## Getting Started

### Install Dependencies

```bash
npm install
```

### Development Mode

Run the development server with hot reload:

```bash
npm run dev
```

Open [http://localhost:3000](http://localhost:3000) to view the app.

**Note:** In development mode, make sure the Go backend is running on port 8080 for API calls to work.

### Production Build

Build the static export for production:

```bash
npm run build
```

This creates an optimized static export in the `out/` directory that can be served by the Go backend.

### Environment Variables

Create a `.env.local` file with:

```
NEXT_PUBLIC_API_URL=http://localhost:8080
```

## Project Structure

```
frontend/
├── app/                    # Next.js app directory
│   ├── page.tsx            # Main application page
│   ├── layout.tsx          # Root layout with metadata
│   └── globals.css         # Global styles and Tailwind imports
├── components/             # React components
│   ├── StatusCard.tsx      # Current work status display
│   ├── MonthlyStatsCard.tsx# Monthly statistics display
│   ├── CheckInSection.tsx  # Check-in form
│   ├── ReCheckInSection.tsx# Re-check-in form
│   ├── CheckOutSection.tsx # Check-out form
│   └── ConfigSection.tsx   # Configuration settings
├── lib/                    # Utility libraries
│   ├── api.ts              # API client for backend communication
│   ├── types.ts            # TypeScript type definitions
│   └── utils.ts            # Shared utility functions
└── out/                    # Production build output (created after build)
```

## Key Features

- **Type-Safe API Client**: All API calls are typed with TypeScript interfaces
- **Real-time Updates**: Status refreshes every 30 seconds
- **Web Notifications**: Browser notifications for check-out reminders
- **Responsive Design**: Works on desktop and mobile devices
- **Modern UI**: Clean interface built with Tailwind CSS

## Integration with Go Backend

The frontend is built as a static export and served by the Go backend at port 8080. The backend serves:
- Static files from `frontend/out/` at the root path
- API endpoints under `/api/*`

## Development Workflow

1. Make changes to components or pages
2. Test in development mode: `npm run dev`
3. Build for production: `npm run build`
4. Run the Go backend to serve the built frontend: `go run main.go`
