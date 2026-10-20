<!--
INSTRUCTIONS FOR THIS FILE (user and AI: read before editing. Keep this header.)

WHAT THIS FILE IS: the end state this project exists to reach, in plain
words. TODO.md and PLAN.md serve this file; work continues until the end
goal is reached. When David restates or extends the end goal, update this
file the same day and let PLAN/TODO follow.

EXACT STRUCTURE:
- "## The end state": the finished project described as the experience of
  the person using it, start to finish.
- "## How we will know it is reached": observable conditions, each one
  checkable, tied to the TODO item that proves it where one exists.
- "## Not in the end state": explicit exclusions David has set, so the
  goal cannot drift.

RULES:
- Sourced from David's stated vision, the repo README and features, and
  what he has talked about in the record. Never invented. Where the
  record is thin, say the end state at the level the record supports.
- No secrets, passwords, or tokens in this file, ever.
-->
# ENDGOAL: Mimir

The end state this project exists to reach, in plain words. TODO.md
and PLAN.md are how we get here; this file is how we know we are
done. Work continues until this end goal is reached. When David
restates or extends the end goal, this file is updated the same day
and PLAN/TODO follow. Sourced from David's stated vision of
2026-09-30, his goal file, and the repo record.

## The end state
A person with a project to build installs Mimir and hands it the
work. Mimir guides them through planning: what the thing is, what
it must do, what already exists. A repository is created or linked
from their forge account, competitors and features are researched,
and the architecture is worked out with them. That is the
human-guided part, and it ends when they approve the plan.

From approval, Mimir runs on its own. It works the 10 stages
(discover, contract, repo, research, architecture, build, verify,
review, ship, handover) as an inspectable graph it can be watched
and steered through, start to finish, with nobody sitting over it.
It writes the code, verifies it against hard gates, and stops for a
human at exactly two places: verify and ship. Gate failures feed
back into the work until the work passes or the run escalates.

Everything Mimir reads along the way, the user's uploads, the repos
it ingests, the research it gathers, lands in a typed knowledge
mind map instead of sitting static in a folder. The map feeds
planning, research, building, and verification on this run and
every later one. When two facts contradict, the map flags it and
the user resolves it at the plan gate; nothing is silently
overwritten.

The harness is for everyone, not a private tool for David: the same
install that runs his projects runs anyone's, on their own machine,
against their own forge accounts and their own choice of model.

## How we will know it is reached
- A new user can install Mimir, register a forge account through
  the OAuth flow, link or create a repository, and start a run
  without David's help (today: OAuth registration is stalled,
  TODO T-6.1).
- A run with an approved plan goes from first stage to shipped
  code with no human watching, and the run itself writes the
  report artifact proving it (today: RunUnattended has no
  production caller, TODO T-2.2; no live end-to-end run has
  happened yet, TODO T-2.3).
- All 10 stages execute as the designed run graph with typed done
  gates at every stage and hard human gates at verify and ship
  (today: the graph is designed, not implemented, TODO T-6.4; the
  shipped engine runs a 7-phase SDLC inside runs).
- Uploaded material demonstrably lands in the mind map through the
  full pipeline (parse, extraction, deterministic dedupe, merge
  with contradiction flagging, integration receipt) and is used by
  later stages and later runs (today: GitIngest and the map exist
  in the Cortex, TODO T-0.4; the pipeline to spec is TODO T-6.3).
- Contradictions in the map surface to the user at the plan gate
  and are resolved there, never silently overwritten (trust rule
  from the goal file).
- Engine updates refuse unsigned or badly signed manifests
  (today: verification code shipped; no release key exists yet,
  TODO Phase 3).
- The repository history carries no live credential (today: F22 is
  open, TODO Phase 1).

## Not in the end state
- Mimir is a standalone project. It is NOT part of AlphaFlux, and
  the end state does not fold it into the AlphaFlux platform.
- Shipping is never fully hands-off: the human gates at verify and
  ship stay human in the end state. Unleashed execution begins at
  plan approval, not before it.
- Knowledge is never silently overwritten: contradiction resolution
  belongs to the user, not the merge.
