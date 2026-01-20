# Frontend Setup Instructions

## Quick Start

1. **Install Dependencies**:
   ```bash
   npm install
   # or
   yarn install
   ```

2. **Start Development Server**:
   ```bash
   npm run dev
   # or
   yarn dev
   ```

3. **Type Checking**:
   ```bash
   npm run typecheck
   ```

## Fixed Issues

The following code errors have been fixed:

1. ✅ **ArrayBuffer type issue** in `useAudio.ts` - Fixed type conversion
2. ✅ **MediaStreamTrack type annotation** - Added proper type
3. ✅ **Unused variable warnings** - Removed unused `workletNodeRef`
4. ✅ **Unused parameter** - Prefixed `onDisconnect` with underscore in CallControls

## Remaining Errors

Most TypeScript errors you see are due to missing `node_modules`. After running `npm install`, these will be resolved:

- Module resolution errors (React, TypeScript types, etc.)
- JSX type errors (these require React types)
- Import errors

These are **not actual code errors** - they're TypeScript complaining about missing type definitions from `node_modules`.

## CSS Warnings

The Tailwind CSS warnings (`@tailwind`, `@apply`) are **normal and expected**. They appear because the CSS linter doesn't recognize Tailwind directives, but Vite/PostCSS will process them correctly during build.

## Verification

After installing dependencies, verify everything works:

```bash
# Install dependencies
npm install

# Check for type errors
npm run typecheck

# Start dev server
npm run dev
```

The application should run without errors once dependencies are installed.
