# Mímir - UX Specification

**BMAD Artifact - Phase 6 (Sally, UX Designer)**
Date: 2026-07-23 | Inputs: PRD (E12), Architecture (section 9), F9.1, F24
Scope: the primary web/desktop GUI, the TUI, and (later) mobile.

---

## 1. Design Principles

- **Calm control:** the user always knows what the agent is doing and can pause/steer it.
- **One glance:** status, progress, and pending approvals are always visible.
- **Keyboard-first, mouse-friendly:** everything reachable by keyboard; the GUI adds
  comfort, not dependency.
- **Accessible:** WCAG AA, full keyboard nav, semantic markup (per components.build).
- **Themeable:** design tokens; light + dark; respects system preference.

---

## 2. Primary GUI Layout (F9.1)

A four-region shell plus a static top bar:

```
+================================================================+
|  TOP BAR (static)                                              |
|  [Mimir] [project v] [agent v]  [search......]  [status] [you] |
+------+===========================================+-------------+
| LEFT |                                           | QUICK-LAUNCH|
| NAV  |          CENTER WORKSPACE                 | RAIL (static|
| (col-|  (files / editor / terminal / spec /      |  far right) |
| laps-|   tasks / live preview - tabbed)          |  [tools]    |
| ible)|                                           |  [agents]   |
|      |                                           |  [commands] |
| proj |                                           |  [favorites]|
| cortex|                                          |             |
| sess |                                           |             |
| agents|                                          |             |
| skills|                                          |             |
| sets |                                           |             |
+------+============================+==============+-------------+
                                   |  RIGHT CHAT PANEL          |
                                   |  (talk to the LLMs:        |
                                   |   stream, tool calls,      |
                                   |   approvals, to-do list)   |
                                   +----------------------------+
```

- **Top bar (static):** app title, project switcher, agent switcher, global search,
  connection/status indicator (daemon + model + sandbox), account/settings.
- **Left nav (collapsible):** projects, Cortex (knowledge), sessions, agents, skills,
  settings. Collapses to an icon rail; fully hidden for focus mode.
- **Center workspace (tabbed):** files, code editor (diff view), terminal, spec/tasks,
  and the **live preview** (F24). Tabs are drag-reorderable.
- **Right chat panel:** the conversation with the agent - streaming tokens, tool-call
  cards (expandable), inline **approval prompts**, and the **to-do list** (the small
  model's working memory + the user's progress view). Resizable; can pop out.
- **Quick-launch rail (static, far right):** one-click favorites - pinned tools, agents,
  slash commands. Always visible.

---

## 3. Key Screens / Views

### 3.1 Chat (right panel)
- Streaming message bubbles; reasoning collapsible.
- **Tool-call cards:** show tool name + args; expand for full output; status badge
  (running/done/failed).
- **Approval prompt:** inline card with the proposed action, risk level, and
  Approve / Deny / Always-allow-this buttons (driven by the policy gate).
- **To-do list widget:** the current task's ordered todos with status; click to jump to
  the relevant file/preview. Live-updates as the agent works.

### 3.2 Cortex / Knowledge (left nav -> Cortex)
- The Well (Mímisbrunnr): browse neurons by kind (source/note/concept/memory/skill) and
  layer; search (full-text + semantic); a graph view of Yggdrasil (neurons + synapses).
- **Add knowledge:** drop a repo URL (gitmcp), YouTube link, web page, PDF, or audio -
  the Gjallarhorn ingestion pipeline processes it into neurons.
- Memory view: engrams with decay/importance; the consolidation status.

### 3.3 Live Preview + Annotation (F24)
- A browser pane in the center workspace showing the app/site the agent built
  (served via the sandbox port routing at localhost:N).
- **Annotation mode:** a markup canvas overlay - draw arrows, boxes, freehand, and text
  on top of the preview.
- **Send back:** "Send to agent" packages the annotated screenshot (image + markup text)
  as multimodal feedback into the chat; the agent applies the changes; re-preview.

### 3.4 Sessions
- List of sessions per project; resume/fork/delete; search across sessions.

### 3.5 Agents
- Configured agents (personas: SOUL/USER/IDENTITY), their model, mode, and tool policy.

### 3.6 Settings & Privacy
- Providers + keys (BYOK), model routing, default model.
- **Privacy:** telemetry toggle (on by default; one-click off), what's collected, data
  export/delete.
- Permissions: per-tool/per-command allow/ask/deny; sandbox defaults.

---

## 4. TUI Layout (Bubble Tea)

- Three panes: **sessions/nav** (left), **chat + tool calls** (center), **status** (right
  rail: model, sandbox, to-do progress).
- Leader-key commands (`Ctrl+x`); command palette; inline approval prompts; streaming.
- The to-do list renders as a live checklist in the status rail.
- Parity with the GUI for all core actions (chat, approve, preview-as-URL, settings).

---

## 5. Mobile (later - F25)

- Single-column: chat-first, with a bottom sheet for the to-do list + approvals.
- Push notifications for approvals/completion; tap to approve from the lock screen.
- Voice input; the same sessions as desktop (portable sessions over the tunnel).

---

## 6. Design System

- **Components:** composable, accessible (Radix-style primitives), copy-in ownership
  (components.build philosophy). Buttons, cards, dialogs, tabs, trees, editors, toasts.
- **Tokens:** color, spacing, radius, typography, elevation; light + dark themes.
- **Motion:** minimal, purposeful (streaming, transitions); respect reduced-motion.
- **Empty/loading/error states** defined for every view.

---

## 7. Key Interaction Flows

- **First run:** install -> onboarding (pick provider or detect Ollama) -> open project
  -> first prompt -> approve a change -> see the diff -> done.
- **Approval:** agent hits a gated action -> approval card in chat (+ push on mobile) ->
  user decides -> agent proceeds or stops.
- **Preview feedback:** agent builds -> preview opens -> user annotates -> send back ->
  agent revises -> re-preview.
- **Goal mode:** user states a goal -> agent shows the game plan -> converts to the to-do
  list -> works item-by-item (visible in the to-do widget) -> verifies -> reports done.

---

## 8. Open UX Questions

- Default theme (dark?) and accent color (forge/ember vs neural/teal)?
- Editor: embed Monaco/CodeMirror, or hand off to the user's editor via ACP?
- Graph view library for Yggdrasil (e.g. a force-directed canvas)?

Next: hand to **John** for story breakdown; the GUI shell (E12) and to-do widget are
designed here and ready to spec into stories.
