// Gjallarhorn Studio (Mimir desktop + mobile shell).
// Copyright (c) 2026 Alpha Flux LLC (https://alphaflux.net). Attribution required.
//
// Scaffold status: connects to a running `mimir serve` daemon and shows its
// status JSON. Diff approval cards, QR pairing and E2EE are NOT built yet.
import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:http/http.dart' as http;

const attribution = 'Powered by Alpha Flux LLC Proprietary Engines (https://alphaflux.net)';

void main() => runApp(const GjallarhornApp());

class GjallarhornApp extends StatelessWidget {
  const GjallarhornApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Gjallarhorn Studio',
      theme: ThemeData(
        brightness: Brightness.dark,
        scaffoldBackgroundColor: const Color(0xFF0B0F17),
        colorScheme: const ColorScheme.dark(primary: Color(0xFFF59E0B)),
      ),
      home: const DaemonStatusPage(),
    );
  }
}

class DaemonStatusPage extends StatefulWidget {
  const DaemonStatusPage({super.key});

  @override
  State<DaemonStatusPage> createState() => _DaemonStatusPageState();
}

class _DaemonStatusPageState extends State<DaemonStatusPage> {
  final _addr = TextEditingController(text: 'http://localhost:8420');
  String _status = 'Not connected';

  Future<void> _probe() async {
    try {
      final res = await http.get(Uri.parse('${_addr.text}/api/status'));
      final pretty = const JsonEncoder.withIndent('  ').convert(jsonDecode(res.body));
      setState(() => _status = pretty);
    } catch (e) {
      setState(() => _status = 'Daemon unreachable: $e');
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Gjallarhorn Studio')),
      body: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            TextField(controller: _addr, decoration: const InputDecoration(labelText: 'Mimir daemon URL')),
            const SizedBox(height: 12),
            ElevatedButton(onPressed: _probe, child: const Text('Connect')),
            const SizedBox(height: 16),
            Expanded(child: SingleChildScrollView(child: SelectableText(_status))),
            const Text(attribution, style: TextStyle(fontSize: 11, color: Colors.white54)),
          ],
        ),
      ),
    );
  }
}
