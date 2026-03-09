# CLAUDE.md — Révise Mieux

## Projet

Révise Mieux est un SaaS éducatif qui transforme des photos de cahier en assistant de révision pour collégiens. Le dépôt contient pour l'instant uniquement les spécifications (PRD + 185 critères d'acceptation). L'implémentation n'a pas encore démarré.

## Documents clés

- `docs/PRD.md` — PRD complet (personas, pipeline, architecture, modèle de données, algorithmes, SLA)
- `docs/MVP-scope.md` — Classification des 185 ACs et périmètre Lot 0 (55 ACs)
- `docs/ac/Z1.md` à `docs/ac/Z8.md` — Critères d'acceptation détaillés par zone (format Given/When/Then)

## Concepts métier essentiels

- **Mastery** : machine à états UNKNOWN → FRAGILE → OK → SOLID avec régressions sur échec
- **Répétition espacée** : transition OK → SOLID requiert 2 réussites espacées de 24h minimum
- **Pipeline J0** : upload photo → segmentation → OCR → structuration LLM → génération items
- **HITL** : validation humaine (parent/admin) des items générés par le LLM
- **Lazy generation** : items générés à la demande, cache 24h, invalidation sur validation ou CRUD exam
- **Lot 0** : 55 ACs pré-MVP pour usage local père-fils sur 4 chapitres pilotes
