<!--
.. title: Intelligent Presenter Documentation
.. slug: index
.. date: 2025-08-18
.. tags: documentation, intelligent-presenter
.. category: 
.. link: 
.. description: Complete documentation for the Intelligent Presenter project
.. type: text
-->

# Intelligent Presenter Documentation

Welcome to the complete documentation for the **Intelligent Presenter** project - an AI-powered HTML slide presentation system with integrated project management features.

## Project Overview

Intelligent Presenter is a sophisticated web application that generates interactive HTML presentations using AI services and integrates with Backlog project management. The system provides:

- **AI-Powered Slide Generation**: Automated content creation using OpenAI or AWS Bedrock
- **Project Integration**: Direct integration with Backlog for project data visualization  
- **Multi-language Support**: Text-to-speech narration in multiple languages
- **Real-time Updates**: WebSocket-based live slide delivery
- **Responsive Design**: Mobile-friendly presentation interface

## Architecture

The project consists of several key components:

### Backend (Go)
- **Main Server**: RESTful API built with Gin framework
- **Authentication**: OAuth2 integration with Backlog
- **AI Services**: Integration with OpenAI and AWS Bedrock
- **MCP Integration**: Model Context Protocol for external data sources
- **Speech Services**: Text-to-speech functionality with multiple providers

### Frontend (TypeScript/Vue)
- **Vue 3 Application**: Modern reactive UI framework
- **Pinia State Management**: Centralized state handling
- **Real-time Communication**: WebSocket integration
- **Slide Processing**: Markdown to HTML conversion with Slidev
- **Audio Integration**: Synchronized narration playback

### Supporting Services
- **Speech Server**: Go-based TTS service
- **MCP Servers**: Backlog integration and external data access
- **Audio Processing**: Multi-provider TTS with caching

## Quick Navigation

### [API Documentation](/pages/api-documentation/)
Complete API reference for all endpoints, WebSocket connections, and data structures.

### [Backend Documentation](/pages/backend-docs/)
Go package documentation covering all services, handlers, and business logic.

### [Frontend Documentation](/pages/frontend-docs/)
TypeScript/Vue component documentation and architecture details.

### [Deployment Guide](/pages/deployment/)
Instructions for deploying the system in production environments.

### [Development Setup](/pages/development/)
Getting started with local development and contributing guidelines.

## Key Features

- **Automated Content Generation**: AI-driven slide creation from project data
- **Interactive Presentations**: HTML-based slides with navigation and animations
- **Project Analytics**: Visual representation of project metrics and progress
- **Multi-modal Output**: Combined visual and audio presentation formats
- **Extensible Architecture**: Plugin-based system for additional data sources

## Technology Stack

- **Backend**: Go, Gin, OAuth2, JWT, WebSockets
- **Frontend**: Vue 3, TypeScript, Vite, Pinia
- **AI Services**: OpenAI GPT, AWS Bedrock (Claude, Titan)
- **Speech**: Multiple TTS providers (Kokoro, MLX, Cloud services)
- **Documentation**: Nikola, TypeDoc, Go doc
- **Testing**: Vitest, Go testing, Playwright E2E

## Getting Started

### For Developers

1. **[Development Setup](/pages/development/)** - Set up your local development environment
2. **[API Documentation](/pages/api-documentation/)** - Understand the system interfaces  
3. **[Architecture Overview](/pages/backend-docs/)** - Learn about the system design
4. **[Frontend Guide](/pages/frontend-docs/)** - Explore the user interface components

### Documentation Quick Links

* 🔗 **[Live TypeScript API Documentation](/files/typescript-docs/index.html)** - Generated from source code
* 📄 **[Go Package Documentation](/files/go-docs.txt)** - Backend package reference

---

This documentation is generated automatically using Nikola + TypeDoc + Go doc integration.