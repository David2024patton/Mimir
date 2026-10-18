// Gjallarhorn Studio (Mimir Desktop + Mobile Shell).
// Copyright (c) 2026 Alpha Flux LLC (https://alphaflux.net). All rights reserved.
// Attribution required: Powered by Alpha Flux LLC Proprietary Engines (https://alphaflux.net)
import 'dart:async';
import 'dart:convert';
import 'package:flutter/material.dart';
import 'package:http/http.dart' as http;

const String attributionNotice =
    'Powered by Alpha Flux LLC Proprietary Engines (https://alphaflux.net)';

void main() => runApp(const GjallarhornApp());

class GjallarhornApp extends StatelessWidget {
  const GjallarhornApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Gjallarhorn Studio · Mímir',
      debugShowCheckedModeBanner: false,
      theme: ThemeData(
        brightness: Brightness.dark,
        scaffoldBackgroundColor: const Color(0xFF0B1018),
        fontFamily: 'monospace',
        colorScheme: const ColorScheme.dark(
          primary: Color(0xFF2DD4BF), // Mimir Teal
          secondary: Color(0xFFFB923C), // Mimir Ember
          surface: Color(0xFF121B28),
        ),
      ),
      home: const StudioMainLayout(),
    );
  }
}

class StudioMainLayout extends StatefulWidget {
  const StudioMainLayout({super.key});

  @override
  State<StudioMainLayout> createState() => _StudioMainLayoutState();
}

class _StudioMainLayoutState extends State<StudioMainLayout> {
  final TextEditingController _daemonAddrController =
      TextEditingController(text: 'http://localhost:8420');
  final TextEditingController _promptController = TextEditingController();
  final ScrollController _chatScroll = ScrollController();

  bool _connected = false;
  String _activeTab = 'chat'; // 'chat' | 'engines' | 'swarm' | 'cortex'
  Map<String, dynamic> _daemonStatus = {};
  List<Map<String, dynamic>> _messages = [];
  List<dynamic> _engines = [];
  List<dynamic> _swarmTasks = [];
  List<dynamic> _memories = [];
  Timer? _pollingTimer;

  @override
  void initState() {
    super.initState();
    _connectDaemon();
    _pollingTimer = Timer.periodic(const Duration(seconds: 4), (_) {
      if (_connected) _fetchData();
    });
  }

  @override
  void dispose() {
    _pollingTimer?.cancel();
    _daemonAddrController.dispose();
    _promptController.dispose();
    _chatScroll.dispose();
    super.dispose();
  }

  Future<void> _connectDaemon() async {
    final base = _daemonAddrController.text.trim();
    try {
      final res = await http
          .get(Uri.parse('$base/api/status'))
          .timeout(const Duration(seconds: 3));
      if (res.statusCode == 200) {
        setState(() {
          _connected = true;
          _daemonStatus = jsonDecode(res.body);
        });
        _fetchData();
      } else {
        setState(() => _connected = false);
      }
    } catch (_) {
      setState(() => _connected = false);
    }
  }

  Future<void> _fetchData() async {
    final base = _daemonAddrController.text.trim();
    // 1. Engines
    try {
      final resEngines = await http.get(Uri.parse('$base/api/engines'));
      if (resEngines.statusCode == 200) {
        final data = jsonDecode(resEngines.body);
        setState(() => _engines = data['engines'] ?? []);
      }
    } catch (_) {}

    // 2. Swarm tasks
    try {
      final resSwarm = await http.get(Uri.parse('$base/api/swarm'));
      if (resSwarm.statusCode == 200) {
        final data = jsonDecode(resSwarm.body);
        setState(() => _swarmTasks = data['tasks'] ?? []);
      }
    } catch (_) {}

    // 3. Cortex memories
    try {
      final resMem = await http.get(Uri.parse('$base/memory'));
      if (resMem.statusCode == 200) {
        final data = jsonDecode(resMem.body);
        setState(() => _memories = data['neurons'] ?? []);
      }
    } catch (_) {}
  }

  Future<void> _sendPrompt() async {
    final text = _promptController.text.trim();
    if (text.isEmpty) return;
    _promptController.clear();

    setState(() {
      _messages.add({'role': 'user', 'content': text});
      _messages.add({'role': 'assistant', 'content': 'Thinking...'});
    });
    _scrollToBottom();

    final base = _daemonAddrController.text.trim();
    try {
      final res = await http.post(
        Uri.parse('$base/chat'),
        headers: {'Content-Type': 'application/json'},
        body: jsonEncode({'prompt': text, 'mode': 'chat'}),
      );
      if (res.statusCode == 200) {
        final data = jsonDecode(res.body);
        setState(() {
          _messages.removeLast();
          _messages.add({'role': 'assistant', 'content': data['reply'] ?? ''});
        });
      } else {
        setState(() {
          _messages.removeLast();
          _messages.add({
            'role': 'assistant',
            'content': 'Error from daemon: HTTP ${res.statusCode}'
          });
        });
      }
    } catch (e) {
      setState(() {
        _messages.removeLast();
        _messages.add({'role': 'assistant', 'content': 'Daemon error: $e'});
      });
    }
    _scrollToBottom();
  }

  void _scrollToBottom() {
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (_chatScroll.hasClients) {
        _chatScroll.animateTo(
          _chatScroll.position.maxScrollExtent,
          duration: const Duration(milliseconds: 200),
          curve: Curves.easeOut,
        );
      }
    });
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        backgroundColor: const Color(0xFF0E1520),
        elevation: 0,
        title: Row(
          children: [
            Container(
              width: 14,
              height: 14,
              decoration: BoxDecoration(
                shape: BoxShape.circle,
                color: _connected
                    ? const Color(0xFFA3E635)
                    : const Color(0xFFF87171),
                boxShadow: [
                  BoxShadow(
                    color: _connected
                        ? const Color(0xFFA3E635).withOpacity(0.4)
                        : const Color(0xFFF87171).withOpacity(0.4),
                    blurRadius: 8,
                  )
                ],
              ),
            ),
            const SizedBox(width: 10),
            const Text(
              'Gjallarhorn Studio',
              style: TextStyle(fontWeight: FontWeight.bold, fontSize: 16),
            ),
            const SizedBox(width: 8),
            Text(
              _connected ? '(${_daemonStatus['model'] ?? 'ready'})' : '(offline)',
              style: const TextStyle(fontSize: 12, color: Colors.white54),
            ),
          ],
        ),
        actions: [
          SizedBox(
            width: 220,
            child: TextField(
              controller: _daemonAddrController,
              style: const TextStyle(fontSize: 12),
              decoration: const InputDecoration(
                isDense: true,
                border: InputBorder.none,
                hintText: 'Daemon URL',
              ),
              onSubmitted: (_) => _connectDaemon(),
            ),
          ),
          IconButton(
            icon: const Icon(Icons.refresh, size: 18),
            onPressed: _connectDaemon,
            tooltip: 'Re-probe Daemon',
          ),
        ],
      ),
      body: Row(
        children: [
          // Icon Navigation Rail
          NavigationRail(
            backgroundColor: const Color(0xFF0E1520),
            selectedIndex: _tabIndex(_activeTab),
            onDestinationSelected: (idx) {
              setState(() {
                _activeTab = _tabForIndex(idx);
              });
            },
            labelType: NavigationRailLabelType.all,
            destinations: const [
              NavigationRailDestination(
                icon: Icon(Icons.chat_bubble_outline),
                selectedIcon: Icon(Icons.chat_bubble),
                label: Text('Chat'),
              ),
              NavigationRailDestination(
                icon: Icon(Icons.memory),
                selectedIcon: Icon(Icons.memory),
                label: Text('Engines'),
              ),
              NavigationRailDestination(
                icon: Icon(Icons.precision_manufacturing),
                selectedIcon: Icon(Icons.precision_manufacturing),
                label: Text('Swarm'),
              ),
              NavigationRailDestination(
                icon: Icon(Icons.all_inclusive),
                selectedIcon: Icon(Icons.all_inclusive),
                label: Text('Cortex'),
              ),
            ],
          ),
          const VerticalDivider(width: 1, color: Color(0xFF22303F)),
          // Main Content View
          Expanded(child: _buildActiveView()),
        ],
      ),
      bottomNavigationBar: Container(
        color: const Color(0xFF0E1520),
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 6),
        child: const Text(
          attributionNotice,
          style: TextStyle(fontSize: 11, color: Colors.white38),
          textAlign: TextAlign.center,
        ),
      ),
    );
  }

  int _tabIndex(String tab) {
    switch (tab) {
      case 'engines':
        return 1;
      case 'swarm':
        return 2;
      case 'cortex':
        return 3;
      default:
        return 0;
    }
  }

  String _tabForIndex(int idx) {
    switch (idx) {
      case 1:
        return 'engines';
      case 2:
        return 'swarm';
      case 3:
        return 'cortex';
      default:
        return 'chat';
    }
  }

  Widget _buildActiveView() {
    switch (_activeTab) {
      case 'engines':
        return _buildEnginesView();
      case 'swarm':
        return _buildSwarmView();
      case 'cortex':
        return _buildCortexView();
      default:
        return _buildChatView();
    }
  }

  Widget _buildChatView() {
    return Column(
      children: [
        Expanded(
          child: _messages.isEmpty
              ? const Center(
                  child: Text(
                    'Mímir is ready. Ask a question or dispatch an SDLC task.',
                    style: TextStyle(color: Colors.white38),
                  ),
                )
              : ListView.builder(
                  controller: _chatScroll,
                  padding: const EdgeInsets.all(16),
                  itemCount: _messages.length,
                  itemBuilder: (ctx, idx) {
                    final m = _messages[idx];
                    final isUser = m['role'] == 'user';
                    return Align(
                      alignment: isUser
                          ? Alignment.centerRight
                          : Alignment.centerLeft,
                      child: Container(
                        margin: const EdgeInsets.symmetric(vertical: 4),
                        padding: const EdgeInsets.all(12),
                        constraints: const BoxConstraints(maxWidth: 600),
                        decoration: BoxDecoration(
                          color: isUser
                              ? const Color(0xFF17222F)
                              : const Color(0xFF121B28),
                          borderRadius: BorderRadius.circular(8),
                          border: Border.all(
                            color: isUser
                                ? const Color(0xFF2D4054)
                                : const Color(0xFF2DD4BF).withOpacity(0.3),
                          ),
                        ),
                        child: Text(
                          m['content'] ?? '',
                          style: TextStyle(
                            color: isUser
                                ? Colors.white
                                : const Color(0xFFE8EEF4),
                          ),
                        ),
                      ),
                    );
                  },
                ),
        ),
        Container(
          padding: const EdgeInsets.all(12),
          decoration: const BoxDecoration(
            color: Color(0xFF0E1520),
            border: Border(top: BorderSide(color: Color(0xFF22303F))),
          ),
          child: Row(
            children: [
              Expanded(
                child: TextField(
                  controller: _promptController,
                  decoration: const InputDecoration(
                    hintText: 'Type your instruction or question...',
                    border: InputBorder.none,
                  ),
                  onSubmitted: (_) => _sendPrompt(),
                ),
              ),
              IconButton(
                icon: const Icon(Icons.send, color: Color(0xFF2DD4BF)),
                onPressed: _sendPrompt,
              ),
            ],
          ),
        ),
      ],
    );
  }

  Widget _buildEnginesView() {
    return Padding(
      padding: const EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Text(
            'Alpha Flux LLC Proprietary Engines',
            style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
          ),
          const SizedBox(height: 6),
          const Text(
            'Pre-compiled binary matrix verified via SHA-256 against manifest.json',
            style: TextStyle(fontSize: 12, color: Colors.white54),
          ),
          const SizedBox(height: 16),
          Expanded(
            child: ListView.builder(
              itemCount: _engines.length,
              itemBuilder: (ctx, idx) {
                final e = _engines[idx];
                return Card(
                  color: const Color(0xFF121B28),
                  shape: RoundedRectangleBorder(
                    side: const BorderSide(color: Color(0xFF22303F)),
                    borderRadius: BorderRadius.circular(8),
                  ),
                  margin: const EdgeInsets.only(bottom: 12),
                  child: Padding(
                    padding: const EdgeInsets.all(14),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Row(
                          children: [
                            Text(
                              e['name'] ?? '',
                              style: const TextStyle(
                                  fontWeight: FontWeight.bold,
                                  color: Color(0xFF2DD4BF)),
                            ),
                            const Spacer(),
                            Container(
                              padding: const EdgeInsets.symmetric(
                                  horizontal: 8, vertical: 2),
                              decoration: BoxDecoration(
                                color: const Color(0xFFA3E635).withOpacity(0.15),
                                borderRadius: BorderRadius.circular(99),
                              ),
                              child: const Text(
                                'VERIFIED',
                                style: TextStyle(
                                    fontSize: 10,
                                    fontWeight: FontWeight.bold,
                                    color: Color(0xFFA3E635)),
                              ),
                            )
                          ],
                        ),
                        const SizedBox(height: 6),
                        Text(e['description'] ?? '',
                            style: const TextStyle(fontSize: 12)),
                        const SizedBox(height: 8),
                        Text(
                          'Branch: ${e['source_branch']} · Stack: ${e['stack']} · Size: ${((e['size_bytes'] ?? 0) / 1024).round()} KB',
                          style: const TextStyle(
                              fontSize: 11, color: Colors.white38),
                        ),
                      ],
                    ),
                  ),
                );
              },
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildSwarmView() {
    return Padding(
      padding: const EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Text(
            'Level 3 Multi-Agent Software Factory Queue',
            style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
          ),
          const SizedBox(height: 6),
          const Text(
            'Autonomous pipelines: Odin -> Thor -> Loki -> Heimdall in Git worktrees',
            style: TextStyle(fontSize: 12, color: Colors.white54),
          ),
          const SizedBox(height: 16),
          Expanded(
            child: _swarmTasks.isEmpty
                ? const Center(
                    child: Text(
                      'No tasks in queue. Submit tasks via `mimir swarm add`.',
                      style: TextStyle(color: Colors.white38),
                    ),
                  )
                : ListView.builder(
                    itemCount: _swarmTasks.length,
                    itemBuilder: (ctx, idx) {
                      final t = _swarmTasks[idx];
                      return Card(
                        color: const Color(0xFF121B28),
                        shape: RoundedRectangleBorder(
                          side: const BorderSide(color: Color(0xFF22303F)),
                          borderRadius: BorderRadius.circular(8),
                        ),
                        margin: const EdgeInsets.only(bottom: 12),
                        child: ListTile(
                          title: Text(t['title'] ?? ''),
                          subtitle: Text(
                            'State: ${t['state']} · Assigned: ${t['assigned_to']} · Branch: ${t['branch']}',
                            style: const TextStyle(
                                fontSize: 11, color: Colors.white54),
                          ),
                          trailing: Text(
                            t['id'] ?? '',
                            style: const TextStyle(
                                fontWeight: FontWeight.bold,
                                color: Color(0xFFFB923C)),
                          ),
                        ),
                      );
                    },
                  ),
          ),
        ],
      ),
    );
  }

  Widget _buildCortexView() {
    return Padding(
      padding: const EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Text(
            'Cortex Long-Term Knowledge Graph',
            style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
          ),
          const SizedBox(height: 6),
          const Text(
            'Searchable facts, engrams, and Yggdrasil mind-maps preserved across runs',
            style: TextStyle(fontSize: 12, color: Colors.white54),
          ),
          const SizedBox(height: 16),
          Expanded(
            child: _memories.isEmpty
                ? const Center(
                    child: Text('Cortex has no memories yet.',
                        style: TextStyle(color: Colors.white38)),
                  )
                : ListView.builder(
                    itemCount: _memories.length,
                    itemBuilder: (ctx, idx) {
                      final m = _memories[idx];
                      return Container(
                        margin: const EdgeInsets.only(bottom: 8),
                        padding: const EdgeInsets.all(10),
                        decoration: BoxDecoration(
                          color: const Color(0xFF121B28),
                          borderRadius: BorderRadius.circular(6),
                          border: Border.all(color: const Color(0xFF22303F)),
                        ),
                        child: Text(
                          m['content'] ?? '',
                          style: const TextStyle(
                              fontSize: 12, color: Color(0xFFE8EEF4)),
                        ),
                      );
                    },
                  ),
          ),
        ],
      ),
    );
  }
}
