package claude

// GenerateQuestionPrompt is the prompt template for generating questions
const GenerateQuestionPrompt = `あなたは英語学習アシスタントです。英作文練習用の日本語の文を生成してください。

テーマ: %s
難易度: %s

ルール:
- 日本語の文のみを出力してください（それ以外は何も出力しないでください）
- 自然でよく使われる表現にしてください
- 難易度に合わせてください
  - beginner（初級）: シンプルな文法、基本的な語彙
  - intermediate（中級）: 複文、一般的な表現
  - advanced（上級）: 複雑な文法、慣用句、ニュアンスのある表現

日本語の文を出力してください:`

// EvaluateAnswerPrompt is the prompt template for evaluating answers
const EvaluateAnswerPrompt = `あなたは英語学習アシスタントです。ユーザーの英訳を評価してください。

日本語の文: %s
ユーザーの回答: %s

以下の形式で評価してください:
SCORE: [0-100の数値]
MODEL_ANSWER: [あなたの理想的な英訳]
FEEDBACK: [日本語での詳細なフィードバック。文法の訂正、語彙の提案、コメントを含めてください]

正確に評価してください。間違いがあれば指摘し、改善方法を説明してください。`
