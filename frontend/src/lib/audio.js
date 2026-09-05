// Síntesis de Audio Nativa con Web Audio API (Cero dependencias de archivos externos)
export function playChime() {
  try {
    const AudioContext = window.AudioContext || window.webkitAudioContext;
    if (!AudioContext) return;
    const ctx = new AudioContext();

    const notes = [523.25, 659.25, 783.99]; // Acorde mayor C5 - E5 - G5 elegante
    const now = ctx.currentTime;

    notes.forEach((freq, index) => {
      const osc = ctx.createOscillator();
      const gain = ctx.createGain();

      osc.type = 'sine';
      osc.frequency.setValueAtTime(freq, now + index * 0.08);

      gain.gain.setValueAtTime(0.001, now + index * 0.08);
      gain.gain.exponentialRampToValueAtTime(0.2, now + index * 0.08 + 0.02);
      gain.gain.exponentialRampToValueAtTime(0.0001, now + index * 0.08 + 0.9);

      osc.connect(gain);
      gain.connect(ctx.destination);

      osc.start(now + index * 0.08);
      osc.stop(now + index * 0.08 + 0.9);
    });
  } catch {
    // Ignorar si el navegador bloquea audio antes de interacción
  }
}
