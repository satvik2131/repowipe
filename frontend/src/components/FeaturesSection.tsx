import { CheckCircle } from "lucide-react";

const features = [
  {
    title: "Bulk Deletion",
    description: "Delete multiple repositories at once, saving you significant time and effort."
  },
  {
    title: "User-Friendly Interface", 
    description: "An intuitive and clean interface makes selecting and deleting repositories a breeze."
  },
  {
    title: "Secure & Reliable",
    description: "We use GitHub's official API, ensuring your account and data are always secure."
  },
  {
    title: "Repository Filtering",
    description: "Quickly find the repositories you're looking for with powerful search and filtering options."
  },
  {
    title: "Safety Confirmation",
    description: "A final confirmation step prevents accidental deletion of important repositories."
  },
  {
    title: "No Installation Needed",
    description: "RepoWipe is a web-based tool, so there's nothing to download or install."
  }
];

export const FeaturesSection = () => {
  return (
    <section className="py-20 sm:py-24 bg-background" id="features">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="text-center mb-16">
          <h2 className="text-3xl sm:text-4xl font-semibold text-foreground mb-4 font-space">
            Powerful Features for Effortless Management
          </h2>
          <p className="text-lg text-muted-foreground max-w-3xl mx-auto font-space">
            RepoWipe is designed to be simple yet powerful, giving you the tools you need to manage your repositories efficiently.
          </p>
        </div>
        
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
          {features.map((feature, index) => (
            <div key={index} className="flex items-start space-x-4">
              <div className="text-primary text-2xl flex-shrink-0 mt-1">
                <CheckCircle className="h-6 w-6" />
              </div>
              <div>
                <h3 className="text-xl font-bold text-foreground font-space">
                  {feature.title}
                </h3>
                <p className="text-muted-foreground mt-1 font-space">
                  {feature.description}
                </p>
              </div>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
};