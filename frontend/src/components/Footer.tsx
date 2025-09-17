export const Footer = () => {
  return (
    <footer className="bg-secondary border-t border-border">
      <div className="max-w-7xl mx-auto py-12 px-4 sm:px-6 lg:px-8">
        <div className="flex flex-col sm:flex-row justify-between items-center gap-4">
          <p className="text-muted-foreground text-sm font-space text-center sm:text-left">
            © 2024 RepoWipe. All rights reserved.
          </p>
          <p className="text-muted-foreground text-sm font-space text-center sm:text-right">
            A tool for developers, by developers.
          </p>
        </div>
      </div>
    </footer>
  );
};