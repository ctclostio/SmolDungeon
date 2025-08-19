export class SeededRNG {
  private seed: number;

  constructor(seed: number) {
    this.seed = seed;
  }

  next(): number {
    this.seed = (this.seed * 9301 + 49297) % 233280;
    return this.seed / 233280;
  }

  roll(sides: number): number {
    return Math.floor(this.next() * sides) + 1;
  }

  rollD20(): number {
    return this.roll(20);
  }

  rollD6(): number {
    return this.roll(6);
  }

  clone(): SeededRNG {
    return new SeededRNG(this.seed);
  }
}