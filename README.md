# Full-stack web app built only with AI

A university assignment where the rule was that the whole application had to be written by AI. My job was to plan it, prompt the AI, run and test what it produced, and document what worked and what didn't.

- **Backend:** Go
- **Frontend:** React + Vite
- **AI tools:** <!-- e.g. Claude, ChatGPT, Copilot: list the ones you actually used -->

> **Status:** finished as a course assignment. The backend and the frontend each work on their own, but they do not work together (see "What went wrong" below).

## What it does

<!-- Two or three sentences: what the website is for and its main features. -->

## Project structure

```
backend/    Go server
frontend/   React + Vite app
```
<!-- Adjust the folder names if yours are different. -->

## Running it

Backend (needs Go):

```
cd backend
go run .
```

Frontend (needs Node.js):

```
cd frontend
npm install
npm run dev
```

Then open the address Vite prints (by default http://localhost:5173).
<!-- Check these commands against the repo and add any environment variables or database setup it needs. -->

## How I worked with the AI

1. I described the application and its features in plain language.
2. I had the AI generate the backend and the frontend, in separate sessions.
3. I ran the code, reported errors back to the AI and asked for fixes until each part ran.

## What went wrong and what I learned

- **A short description is not enough for a whole application.** With only a rough idea of the design, the AI filled the gaps with its own assumptions, and the result drifted from what I had in mind. It worked much better on small, clearly specified tasks.
- **Without an API contract, the two halves did not fit together.** The backend and the frontend were generated in separate sessions, and the AI could not keep both in its context at the same time. Each side invented its own endpoints and data formats, so both ran on their own but could not talk to each other. Next time I would write the API contract first (for example an OpenAPI file) and give it to the AI in every session.
- **The output still has to be checked by a person.** Code that compiles and runs is not necessarily correct. Testing it myself was the only way to find out where it was wrong.
