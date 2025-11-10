---
title: Authentication and Identity
sidebar_label: Auth & Identity
description: AT Protocol authentication, security, and session handling.
sidebar_position: 8
---

# Authentication and Identity

## AT Protocol Authentication

Noteleaf uses AT Protocol's authentication system:

1. **Handle Resolution**: Your handle (e.g., `username.bsky.social`) is resolved to a DID (Decentralized Identifier)
2. **Session Creation**: Authenticate with your app password to create a session
3. **Session Token**: Noteleaf stores the session token for future requests
4. **Token Refresh**: Sessions are refreshed automatically when they expire

## Security Considerations

**Use app passwords**: Never use your main BlueSky password with third-party tools. App passwords can be revoked without affecting your account.

**Token storage**: Session tokens are stored locally in the Noteleaf database. Protect your database file.

**Revocation**: If compromised, revoke the app password at [bsky.app/settings/app-passwords](https://bsky.app/settings/app-passwords).

## Session Management

**Check status**:

```sh
noteleaf pub status
```

**Re-authenticate**:

```sh
noteleaf pub auth
```

Sessions typically last 2-4 hours before requiring refresh. Noteleaf handles refresh automatically, but if authentication fails, run `pub auth` again.
