import { Button } from "@/components/ui/button";
import { Link } from "react-router-dom";

export const HeroSection = () => {
  return (
    <section className="text-center py-20 sm:py-32 bg-gradient-hero">
      <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8">
        <h1 className="text-4xl sm:text-6xl font-bold text-foreground mb-4 leading-tight tracking-tighter font-space">
          Wipe or Move Repos Across Git Hosts.
        </h1>
        <p className="text-lg text-muted-foreground leading-relaxed mb-8 max-w-2xl mx-auto font-space">
          Bulk-delete on GitHub, GitLab, or Bitbucket — or transfer repositories
          any-to-any with history, issues, and PRs/MRs where the APIs allow.
        </p>
        <Link to="/search">
          <Button variant="hero" size="lg" className="font-space">
            Get Started for Free
          </Button>
        </Link>
      </div>
    </section>
  );
};