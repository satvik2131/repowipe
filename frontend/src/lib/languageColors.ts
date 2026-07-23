/** Common GitHub language color map (subset). */
const LANGUAGE_COLORS: Record<string, string> = {
  TypeScript: "#3178c6",
  JavaScript: "#f1e05a",
  Python: "#3572A5",
  Go: "#00ADD8",
  Rust: "#dea584",
  Java: "#b07219",
  HTML: "#e34c26",
  CSS: "#563d7c",
  Vue: "#41b883",
  Ruby: "#701516",
  PHP: "#4F5D95",
  C: "#555555",
  "C++": "#f34b7d",
  "C#": "#178600",
  Swift: "#F05138",
  Kotlin: "#A97BFF",
  Dart: "#00B4AB",
  Shell: "#89e051",
  Dockerfile: "#384d54",
  MDX: "#fcb32c",
  Markdown: "#083fa1",
};

export function languageColor(language: string | null | undefined): string {
  if (!language) return "hsl(var(--muted-foreground))";
  return LANGUAGE_COLORS[language] ?? "#8b949e";
}

export const COMMON_LANGUAGES = [
  "TypeScript",
  "JavaScript",
  "Python",
  "Go",
  "Rust",
  "Java",
  "HTML",
  "CSS",
  "Vue",
  "Ruby",
  "PHP",
  "C",
  "C++",
  "C#",
  "Swift",
  "Kotlin",
  "Dart",
  "Shell",
] as const;
