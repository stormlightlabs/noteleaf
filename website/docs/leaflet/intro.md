---
title: Leaflet.pub Introduction
sidebar_label: Introduction
description: Understand leaflet.pub and how Noteleaf integrates with it.
sidebar_position: 1
---

# Leaflet.pub Introduction

## What is Leaflet.pub?

[Leaflet.pub](https://leaflet.pub) is a decentralized publishing platform built on the AT Protocol (the same protocol that powers BlueSky). It allows you to publish long-form content as structured documents while maintaining ownership and control of your data through decentralized identity.

## AT Protocol and Decentralized Publishing

AT Protocol provides:

**Portable Identity**: Your identity (DID) is separate from any single service. You own your content and can move it between providers.

**Verifiable Data**: All documents are content-addressed and cryptographically signed, ensuring authenticity and preventing tampering.

**Interoperability**: Content published to leaflet.pub can be discovered and consumed by any AT Protocol-compatible client.

**Decentralized Storage**: Data is stored in personal data repositories (PDSs) under your control, not locked in a proprietary platform.

## How Noteleaf Integrates with Leaflet

Noteleaf can act as a leaflet.pub client, allowing you to:

1. **Authenticate** with your BlueSky/AT Protocol identity
2. **Pull** existing documents from leaflet.pub into local notes
3. **Publish** local notes as new leaflet documents
4. **Update** previously published documents with changes
5. **Manage** drafts and published content from the command line

This integration lets you write locally in markdown, manage content alongside tasks and research notes, and publish to a decentralized platform when ready.
