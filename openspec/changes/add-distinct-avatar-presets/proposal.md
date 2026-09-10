# add-distinct-avatar-presets

## Why

The current app uses the same numeric-looking placeholder avatar set for both users and pets. This weakens the AI pet positioning: user identity should feel trustworthy or privacy-friendly, while pet identity should feel warm, expressive, and emotionally engaging.

## What Changes

- Split app avatar assets into separate user and pet preset sets.
- User avatars remain 8 choices: 4 modern human portraits and 4 abstract/privacy-friendly symbols.
- Pet avatars expand to 12 choices: 6 cute companion pets and 6 cartoon/anime identity styles.
- Pet avatar validation expands from 1~8 to 1~12; user avatar validation remains 1~8.

## Impact

- API behavior: pet create/update accepts avatar_id 1~12; user profile remains avatar_id 1~8.
- App: onboarding/profile/contact surfaces render user avatars; pet creation/home/chat render pet avatars.
- Assets: new raster PNG preset avatar directories are bundled by Flutter.
