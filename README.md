# repowipe

repowipe is a web application that allows users to delete multiple Git repositories efficiently and securely.

## Features

- **Bulk Repository Deletion:** Select and delete multiple repositories in one action.
- **Safe Authentication:** Connect your Git provider (e.g., GitHub) securely.
- **Preview & Confirm:** Review selected repositories before deletion.
- **Progress Feedback:** See real-time status of deletion operations.
- **Error Handling:** Get notified if any repository fails to delete.

## Usage

1. **Login:** Authenticate with your Git provider.
2. **Select Repositories:** Browse and select repositories you want to delete.
3. **Review:** Confirm your selection before proceeding.
4. **Delete:** Initiate deletion. Progress and results are displayed.
5. **Logout:** Disconnect your account when done.

## Technologies

- React + Vite frontend
- REST API backend (not included here)
- OAuth for authentication

## Warning

Deleting repositories is **irreversible**. Ensure you have backups before using repowipe.

## Development

Install dependencies and run locally:

```bash
npm install
npm run dev
```

## License

MIT
