import { NextResponse } from 'next/server';
import { fetchMeaning } from '../../../lib/meaning.js';

interface RequestBody {
  word: string;
}

export async function POST(request: Request): Promise<NextResponse> {
  let body: RequestBody;
  try {
    body = (await request.json()) as RequestBody;
  } catch {
    return NextResponse.json({ error: 'リクエストボディが不正です' }, { status: 400 });
  }

  const { word } = body;
  if (!word || typeof word !== 'string') {
    return NextResponse.json({ error: 'word は必須の文字列です' }, { status: 400 });
  }

  const meaning = await fetchMeaning(word.trim());
  return NextResponse.json({ meaning });
}
