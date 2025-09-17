import { Trash2, List, CheckCircle } from "lucide-react";

const steps = [
  {
    icon: <Trash2 className="h-8 w-8" />,
    title: "1. Select Repositories",
    description: "Connect your GitHub account and easily select the repositories you want to remove."
  },
  {
    icon: <List className="h-8 w-8" />,
    title: "2. Review & Confirm", 
    description: "Double-check your selection. There's no going back after this step, so be sure!"
  },
  {
    icon: <CheckCircle className="h-8 w-8" />,
    title: "3. Wipe Clean",
    description: "Confirm the deletion, and watch as RepoWipe securely removes the repositories for you."
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
            A straightforward, three-step process to a cleaner GitHub profile.
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