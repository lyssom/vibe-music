// === intro (4 bars) ===
sound("hh").every(4).fast(2).swing(60)
chord("c4 e4 g4", "2n", 0.5).every(4).swing(60)
note("c5", "8n", 0.4).every(2).swing(60)
bass("c2", "4n", 0.8).every(4)

// === verse (8 bars) ===
sound("hh").every(4).every(2).fast(2)
chord("c4 e4 g4", "4n", 0.5).every(4)
chord("f4 a4 c5", "4n", 0.5).every(4).every(4)
bass("c2", "4n", 0.7).every(4)
note("e5", "8n", 0.4).every(2)
note("g5", "8n", 0.4).every(4).every(2)

// === bridge (8 bars) ===
sound("bd sd").every(4).every(2)
chord("eb4 g4 bb4", "4n", 0.5).every(4)
chord("ab3 db4 f4", "4n", 0.5).every(4).every(4)
bass("eb1", "4n", 0.7).every(4)
note("bb4", "8n", 0.4).every(2)
note("db5", "8n", 0.4).every(4).every(2)

// === verse2 (8 bars) ===
sound("hh").every(4).fast(2).swing(60)
chord("g4 bb4 d5", "4n", 0.5).every(4)
chord("c4 e4 g4", "4n", 0.5).every(4).every(4)
bass("g2", "4n", 0.7).every(4)
note("d5", "8n", 0.4).every(2)
note("f5", "8n", 0.4).every(4).every(2)

// === outro (4 bars) ===
sound("bd sd").every(4).every(2).swing(60)
chord("c4 e4 g4", "2n", 0.6).every(4).swing(60)
chord("c4 e4 g4", "1n", 0.7).every(4).every(4)
bass("c2", "2n", 0.8).every(4)
note("c5", "4n", 0.5).every(4)
note("e5 g5", "8n", 0.4).every(4).every(2)