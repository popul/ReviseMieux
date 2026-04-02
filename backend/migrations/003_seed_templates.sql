-- Migration: 003_seed_templates.sql
-- Seed the 3 difficulty-1 question templates used by session composition.

INSERT INTO templates (id, name, question_type, difficulty, prompt_template, eligibility, grading)
VALUES
  ('GEN.KNOW.FLASH_MCQ',      'Flash MCQ',         'MCQ',          1, 'Vrai ou faux : {{term}}',            '{}', '{}'),
  ('GEN.KNOW.DEF_SHORT',      'Définition courte', 'SHORT_ANSWER', 1, 'Définis : {{term}}',                 '{}', '{}'),
  ('GEN.KNOW.CLOZE_KEYWORDS', 'Cloze mots-clés',   'CLOZE',        1, 'Complète avec les mots-clés : {{masked_term}}', '{}', '{}')
ON CONFLICT (id) DO NOTHING;
