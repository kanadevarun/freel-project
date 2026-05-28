# Freel — Multi-Modal Logistics Operating System

Freel is a next-generation logistics operating system designed to orchestrate Road, Rail, Air, and Sea shipping networks across India under a single digital workspace.

## 🏗️ Project Architecture

```text
freel-project/
├── frontend/        # React + Vite Client Application (Tailwind CSS v4)
└── backend/         # Go API Microservices Architecture (Phase 2)
```

## 🚀 Getting Started

### Prerequisites

* Node.js (v18 or higher)
* npm (v9 or higher)

### Installation & Local Run

1. Navigate to the client directory:
   ```bash
   cd frontend
   ```
2. Install dependencies:
   ```bash
   npm install
   ```
3. Launch the development server:
   ```bash
   npm run dev
   ```

Open [http://localhost:5173](http://localhost:5173) in your browser to view the application.

## 📦 Deployment

The client application is configured to deploy automatically on **Vercel** via GitHub triggers. Every push to the `main` branch initiates a production build and deployment.
