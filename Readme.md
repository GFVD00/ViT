# Voice to Docx & Cross-Platform Voice Assistant

A powerful cross-platform application (Desktop & Android) built with **Go** and the **Fyne** GUI framework. It allows you to record voice, perform offline speech-to-text recognition using AI (**Vosk**), automatically correct text via custom filters, sync data over a local TCP network, and export transcriptions directly into `.docx` documents.

---

## 🚀 Key Features

* **Cross-Platform Voice Recording:** Real-time audio capture using `malgo` (Windows and Android) and `arecord` (Linux).
* **Offline AI Speech-to-Text (STT):** Local Vosk model integration ensures complete privacy without sending audio data to the cloud.
* **Smart Word Replacement:** Automatic text correction and terminology substitution based on a custom configuration file (`word.env`).
* **Word Document Export:** Save transcriptions instantly into formatted `.docx` files with custom font settings (Desktop).
* **TCP Network Sync:** Connect multiple devices across your local network to send/receive text and voice data seamlessly.

---

## 📦 Pre-compiled Binaries (No Go Required!)

If you just want to use the application, **you do not need to install Go or compile the code**. 
You can download ready-to-use binaries (such as `.exe` for Windows, `.apk` for Android, etc.) directly from the [Releases](../../releases) page.

---

## ⚙️ Model Setup (Required for STT)

The application relies on an offline Vosk speech recognition model to function:

1. Download a compatible language model from the official [Vosk Models Page](https://alphacep.com/vosk/models) (e.g., `vosk-model-small-ru-0.22`).
2. Place the extracted folder containing the model **directly in the same directory where the program binary is located**. The application will automatically detect any folder whose name starts with `vosk-model`.

---

## Platform-Specific Notes & Troubleshooting
🪟 Windows (TCP Connection)
If you experience issues connecting your Android device to the desktop app via TCP, it may be because the Windows Firewall is blocking incoming connections. You will need to add an inbound firewall rule to open port 8080 for TCP traffic.

🤖 Android (Microphone Permission)
Before using the application for the first time, make sure to allow microphone access in your device settings, otherwise the voice recording feature will not function properly.

---

## 🛠️ Configuration (Optional)

For automatic word replacement, create a `word.env` file in the same directory as the binary with your custom rules:
```env
WORD_REPLACEMENTS=old_word:new_word;another_word:correction


