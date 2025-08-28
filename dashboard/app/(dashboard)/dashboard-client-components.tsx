"use client";

import { Button } from "@/components/ui/button";
import { ArrowRight } from "lucide-react";

export function ViewDocsButton() {
  return (
    <Button 
      variant="outline" 
      size="sm" 
      onClick={() => window.open('https://github.com/karthik/anybase', '_blank')}
    >
      View Docs
      <ArrowRight className="h-4 w-4 ml-2" />
    </Button>
  );
}