#!/bin/bash
set -euo pipefail

# Build Frontend
cd frontend
pnpm build 
cd ..

# Build Backend
cd backend
go build -ldflags="-s -w" -o ../gen-ai .
cd ..
