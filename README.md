# LLMang

Goal-driven LLM framework for Go. By [Carson](https://carsho.dev).

## What It Does

Organizes LLM queries into Goals, Solutions, and Prompts. Focus on outcomes, not prompts. Built from frustrations in [Lang Goo](https://github.com/carsho/lang-goo).

- **Goals**: Define the result.
- **Solutions**: Ways to get there, with canary testing.
- **Prompts**: Model-specific inputs.

## Features

- Type-safe Go structs.
- Canary testing in production.
- Handles logging, retries, rate limits.
- Optional frontend UI.

## Install

```bash
go get github.com/llmang/llmango
