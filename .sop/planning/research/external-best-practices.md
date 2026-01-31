# External Research: Best Practices for Educational AI & OCR

## 1. OCR for Handwritten Text Recognition

### Vision-Language Models (Recommended Approach)

Modern vision-language models like GPT-4o combine visual understanding with contextual reasoning, dramatically improving OCR accuracy:

- **Contextual reasoning**: When a character is ambiguous (e.g., 'a' vs 'o'), the model considers surrounding words and document context
- **End-to-end processing**: No need for separate text detection and recognition stages
- **Multi-language support**: Works well with French educational content

### Image Quality Best Practices

| Factor | Recommendation |
|--------|----------------|
| Resolution | 300+ DPI minimum |
| Noise reduction | Apply blur filtering for low-quality images |
| Image size | Resize very large images before processing |
| Lighting | Ensure even lighting, avoid shadows |

### Performance Benchmarks

- **Printed text**: 92-98% accuracy achievable with GPT-4o
- **Handwritten text**: 85-92% accuracy depending on legibility
- **Processing time**: < 3 seconds/page with OpenAI Vision API

### Implementation Notes

```
OpenAI Vision API approach:
1. Accept image upload (JPEG, PNG, PDF)
2. Resize/compress if needed for API limits
3. Send to gpt-4o with OCR-focused prompt
4. Extract structured text with formatting preserved
```

**Sources:**
- [OpenAI Vision API Guide](https://platform.openai.com/docs/guides/vision)
- [OCR Benchmark 2026](https://research.aimultiple.com/ocr-accuracy/)
- [What Makes OCR Different in 2025](https://photes.io/blog/posts/ocr-research-trend)

---

## 2. AI Flashcard & Quiz Generation

### Pedagogical Foundation

Research-backed principles for effective study materials:

| Technique | Effectiveness |
|-----------|---------------|
| **Spaced repetition** | 88% test scores vs 78% traditional methods |
| **Active recall** | 80% recall vs 60% for cramming |
| **Source-grounded generation** | Prevents AI hallucination in educational content |

### Best Practices for Card/Quiz Generation

#### Quality Criteria for Flashcards
1. **Recall-based questions**: "What is X?" not just copying definitions
2. **Specific focus**: Target facts, relationships, or processes
3. **Structured Q&A**: Convert explanations into clean question-answer pairs
4. **Chunked complexity**: Break complex concepts into steps

#### Quiz Generation Features
- **Adaptive difficulty**: Easy → Medium → Hard progression
- **Multiple formats**: MCQ, true/false, short answer
- **Knowledge gap tracking**: Identify weak areas

### Prompt Engineering for Educational Content

Key principles:
- **Source-grounding**: Generate only from provided course content (no external knowledge)
- **Structured output**: Use JSON schemas to ensure consistent format
- **Learning objectives**: Align questions with curriculum goals

**Sources:**
- [NotebookLM AI Flashcards for Education](https://thesciencetalk.com/news/notebooklm-ai-flashcards-science-education-2025/)
- [Best AI Quiz Generator 2025](https://scholarly.so/blog/best-ai-quiz-generator-2025-complete-review)
- [Best AI Flashcard Generators 2025](https://www.quizcat.ai/blog/best-ai-flashcard-generators-for-2025)

---

## 3. LLM Structured Output for JSON

### OpenAI Structured Outputs (Recommended)

OpenAI's Structured Outputs feature ensures:
- **Schema adherence**: Model always generates valid JSON matching your schema
- **No missing keys**: Required fields always present
- **Valid enums**: No hallucinated values

### Implementation Pattern

```json
// Example schema for flashcard generation
{
  "type": "object",
  "properties": {
    "cards": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "question": { "type": "string" },
          "answer": { "type": "string" },
          "difficulty": { "enum": ["easy", "medium", "hard"] }
        },
        "required": ["question", "answer", "difficulty"]
      }
    }
  }
}
```

### Performance Benefits

- **QuizGen benchmark**: Structured outputs improve JSON task performance from 37.6% to 98.8%
- **Throughput**: Up to 1.77x improvement over unconstrained generation

**Sources:**
- [OpenAI Structured Outputs](https://platform.openai.com/docs/guides/structured-outputs)
- [Quiz Maker - LLM Structured Outputs Implementation](https://bettersoftware.medium.com/quiz-maker-a-sample-implementation-for-llm-structured-outputs-56846860fa76)
- [StructuredRAG Benchmark](https://arxiv.org/html/2408.11061v1)

---

## 4. Recommendations for Révise mieux MVP

### OCR Strategy
1. Use **OpenAI Vision API (gpt-4o)** for OCR - handles both printed and handwritten French text
2. Implement image preprocessing (resize, quality check)
3. Add manual correction UI for low-confidence extractions

### Content Generation Strategy
1. Use **OpenAI Structured Outputs** for consistent JSON responses
2. Generate from source content only (source-grounded approach)
3. Include difficulty levels in generated content
4. Create clear, recall-based questions

### Prompt Templates (French)
- **Flashcards**: "À partir du cours suivant, génère des fiches de révision avec question/réponse..."
- **Quiz**: "Crée un quiz de {n} questions à choix multiples basé sur ce contenu..."

### Technical Stack Alignment
- Go backend with OpenAI Go SDK
- JSON schemas for request/response validation
- PostgreSQL for storing courses, cards, quizzes
