import { Header } from "@/components/Header";
import { HeroSection } from "@/components/HeroSection";
import { HowItWorksSection } from "@/components/HowItWorksSection";
import { FeaturesSection } from "@/components/FeaturesSection";
import { Footer } from "@/components/Footer";
import { useEffect } from "react";
import { useAppStore } from "@/store/useAppStore";

const Index = () => {
  const checkAuth = useAppStore((state) => state.checkAuth);
  useEffect(() => {
    checkAuth();
  }, [checkAuth]);

  return (
    <div className="min-h-screen flex flex-col overflow-x-hidden bg-background font-space">
      <Header />
      <main className="flex-1">
        <HeroSection />
        <HowItWorksSection />
        <FeaturesSection />
      </main>
      <Footer />
    </div>
  );
};

export default Index;
