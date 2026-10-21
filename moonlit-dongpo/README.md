# 月下东坡 (Moonlit Dongpo) 🌙

An AI chatbot inspired by **Su Shi (苏轼)** and his poem **《水调歌头》**. Ask questions in Chinese or English to learn about the poem, its meaning, and Su Shi's life.

## 🌐 Just want to use the chatbot?

Open the website link shared by the project owner, type your question in the chat box, and press **Send**. **No installation or API key is needed.**

**Website:** _[Add the public website link here after deployment]_

## 💻 Want to run it on your own computer?

1. **Get the project:** On GitHub, click **Code → Download ZIP**, then extract the ZIP. Open the folder containing `main.go` in VS Code.
2. **Install Go:** Download and install Go from [go.dev/dl](https://go.dev/dl/) if you do not already have it.
3. **Get a free-tier Gemini API key (if eligible):** Visit [Google AI Studio](https://aistudio.google.com/apikey), sign in, and create an API key.
4. **Add your key:** Open the `.env` file already prepared in the project. Replace `YOUR_GEMINI_API_KEY` with your own key. Keep these settings:

   ```env
   GEMINI_API_KEY=YOUR_GEMINI_API_KEY
   GEMINI_MODEL=gemini-3.5-flash-lite
   PORT=8081
   ```

   Save the file. `.env.example` is only a sample; you do not need to edit it. If `.env` is missing from a GitHub download, copy `.env.example` and rename the copy to `.env`.
5. **Start the chatbot:** In VS Code, open **Terminal → New Terminal** in the project folder and type:

   ```bash
   go run .
   ```

6. **Open the website:** Visit **http://localhost:8081** in your browser. Keep the terminal open while using the chatbot.

## 💬 How to use

Try asking: **“苏轼为什么写《水调歌头》？”** The AI will answer in the chat box. You can ask follow-up questions too.

## ❓ Common problems

- **Website does not open:** Check that `go run .` is still running and that you opened `http://localhost:8081`.
- **AI says it is busy:** Check your API key in `.env`, restart the server, and look for errors in the VS Code terminal. Free-tier API usage limits may apply.
- **Port 8081 is already in use:** Close any older running copy of the chatbot before starting it again.

## 🔒 Important

**Never share or upload your `.env` file or Gemini API key to GitHub.** Only share `.env.example` as a sample. Visitors using the public website do not need an API key.

---

Built with **Go**, **Google Gemini**, **HTML**, **CSS**, and **JavaScript**.
