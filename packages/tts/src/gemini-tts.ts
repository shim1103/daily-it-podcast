import { GoogleGenAI, Modality } from '@google/genai';
import type { TtsService, Manuscript } from '@daily-it-podcast/core';
import { TtsError } from '@daily-it-podcast/core';

function buildScript(manuscript: Manuscript): string {
  const { opening, topics, closing } = manuscript.body;
  const topicText = topics.map((t) => `${t.title}\n${t.script}`).join('\n\n');
  return `${opening}\n\n${topicText}\n\n${closing}`;
}

export class GeminiTtsService implements TtsService {
  private readonly ai: GoogleGenAI;
  private readonly model = 'gemini-2.5-flash-preview-tts';

  constructor() {
    const apiKey = process.env['GEMINI_API_KEY'];
    if (!apiKey) {
      throw new TtsError('GEMINI_API_KEY が設定されていません');
    }
    this.ai = new GoogleGenAI({ apiKey });
  }

  async synthesize(manuscript: Manuscript): Promise<Buffer> {
    const text = buildScript(manuscript);

    let response;
    try {
      response = await this.ai.models.generateContent({
        model: this.model,
        contents: [{ parts: [{ text }] }],
        config: {
          responseModalities: [Modality.AUDIO],
          speechConfig: {
            voiceConfig: {
              prebuiltVoiceConfig: { voiceName: 'Kore' },
            },
          },
        },
      });
    } catch (err) {
      throw new TtsError('Gemini TTS API 呼び出しに失敗しました', err);
    }

    const part = response.candidates?.[0]?.content?.parts?.[0];
    if (!part?.inlineData?.data) {
      throw new TtsError('Gemini TTS API からの audio データが空です');
    }

    return Buffer.from(part.inlineData.data, 'base64');
  }
}
