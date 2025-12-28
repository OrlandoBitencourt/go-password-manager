# Password Manager Frontend

Modern, secure password manager web interface built with Next.js 14, TypeScript, and Tailwind CSS.

## Features

- **Modern UI/UX** - Clean, intuitive interface inspired by 1Password and Bitwarden
- **Dark Mode** - Automatic dark mode support with manual toggle
- **Responsive Design** - Works seamlessly on desktop, tablet, and mobile
- **Password Generator** - Generate strong, customizable passwords
- **Real-time Search** - Quickly find passwords with instant search
- **Secure Copy** - One-click copy for usernames and passwords
- **Password Strength Indicator** - Visual feedback on password strength
- **Session Management** - Automatic vault locking for security

## Getting Started

### Prerequisites

- Node.js 18+ and npm
- Go backend running on port 8080 (see root README)

### Development

1. Install dependencies:
```bash
npm install
```

2. Set up environment variables:
```bash
cp .env.local.example .env.local
```

Edit `.env.local` and set `NEXT_PUBLIC_API_URL` to your backend URL (default: `http://localhost:8080`)

3. Run the development server:
```bash
npm run dev
```

4. Open [http://localhost:3000](http://localhost:3000) in your browser

### Production Build

```bash
npm run build
npm start
```

### Docker

Build and run with Docker:
```bash
docker build -t password-manager-frontend .
docker run -p 3000:3000 -e NEXT_PUBLIC_API_URL=http://backend:8080 password-manager-frontend
```

Or use Docker Compose from the root directory:
```bash
docker-compose up frontend
```

## Project Structure

```
frontend/
├── app/
│   ├── components/          # Reusable UI components
│   │   ├── Button.tsx
│   │   ├── Input.tsx
│   │   ├── Card.tsx
│   │   ├── Modal.tsx
│   │   ├── PasswordInput.tsx
│   │   ├── VaultSelector.tsx
│   │   ├── PasswordRecordCard.tsx
│   │   ├── AddPasswordModal.tsx
│   │   └── PasswordGeneratorModal.tsx
│   ├── lib/                 # Utilities and helpers
│   │   ├── api.ts          # API client
│   │   └── password-generator.ts
│   ├── types/              # TypeScript type definitions
│   │   └── index.ts
│   ├── globals.css         # Global styles
│   ├── layout.tsx          # Root layout
│   └── page.tsx            # Main page
├── public/                 # Static assets
├── tailwind.config.ts      # Tailwind configuration
├── next.config.js          # Next.js configuration
└── package.json
```

## Technologies

- **Next.js 14** - React framework with App Router
- **TypeScript** - Type-safe development
- **Tailwind CSS** - Utility-first CSS framework
- **Lucide React** - Beautiful icon library
- **Axios** - HTTP client for API requests

## Features in Detail

### Vault Management

- Create new vaults with strong master passwords
- Unlock vaults with master password authentication
- Lock vaults to clear sensitive data from memory
- List all available vaults

### Password Management

- Add new password records
- View and search password records
- Copy usernames and passwords with one click
- Delete password records
- Show/hide password visibility

### Password Generator

- Customizable length (8-64 characters)
- Toggle character types (uppercase, lowercase, numbers, symbols)
- Real-time password generation
- One-click copy to clipboard
- Visual password strength indicator

### Security Features

- Passwords hidden by default
- Auto-lock on logout
- Secure clipboard operations
- No password storage in browser
- HTTPS recommended for production

## Configuration

Environment variables:

- `NEXT_PUBLIC_API_URL` - Backend API URL (default: `http://localhost:8080`)

## Development Tips

- Hot reload is enabled in development mode
- Use TypeScript for type safety
- Follow the existing component structure
- Use Tailwind utility classes for styling
- Icons from `lucide-react` library

## Contributing

1. Follow the existing code style
2. Use TypeScript types for all components
3. Test on different screen sizes
4. Ensure dark mode compatibility
5. Add comments for complex logic

## License

MIT License - see LICENSE file in root directory
