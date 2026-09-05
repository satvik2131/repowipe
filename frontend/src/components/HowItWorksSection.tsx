import { Trash2, List, CheckCircle } from "lucide-react";

const steps = [
  {
    icon: <List className="h-8 w-8" />,
    title: "1. Connect accounts",
    description: "Log in with one provider, then link GitHub, GitLab, and Bitbucket as needed."
  },
  {
    icon: <Trash2 className="h-8 w-8" />,
    title: "2. Wipe or transfer",
    description: "Pick a source host, select repos, then bulk-delete or mirror them to another host."
  },
  {
    icon: <CheckCircle className="h-8 w-8" />,
    title: "3. Track progress",
    description: "Transfers report succeeded, partial, or failed per repo with warnings when metadata can't move."
  }
];

export const HowItWorksSection = () => {
  return (
    <section className="py-20 sm:py-24 bg-secondary" id="how-it-works">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="text-center mb-16">
          <h2 className="text-3xl sm:text-4xl font-semibold text-foreground mb-4 font-space">
            How It Works
          </h2>
          <p className="text-lg text-muted-foreground max-w-3xl mx-auto font-space">
            A straightforward path to a cleaner profile — or a new home for your repos.
          </p>
        </div>
        
        <div className="grid grid-cols-1 md:grid-cols-3 gap-8 text-center">
          {steps.map((step, index) => (
            <div 
              key={index}
              className="bg-background rounded-lg p-8 shadow-card transform hover:-translate-y-2 transition-smooth"
            >
              <div className="flex items-center justify-center h-16 w-16 rounded-full bg-primary/20 text-primary mx-auto mb-6">
                {step.icon}
              </div>
              <h3 className="text-xl font-bold text-foreground mb-2 font-space">
                {step.title}
              </h3>
              <p className="text-muted-foreground font-space">
                {step.description}
              </p>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
};