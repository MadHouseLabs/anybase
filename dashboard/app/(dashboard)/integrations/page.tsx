'use client';

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { 
  ArrowRight, 
  Plug, 
  Sparkles,
  Brain,
} from 'lucide-react';
import Link from 'next/link';

const integrations = [
  {
    id: 'ai-providers',
    title: 'AI Providers',
    description: 'Configure AI providers for embeddings, RAG, and intelligent search capabilities',
    icon: Brain,
    href: '/integrations/ai-providers',
    status: 'available',
    features: ['OpenAI', 'Anthropic', 'Cohere', 'Ollama'],
  },
];

export default function IntegrationsPage() {
  return (
    <div className="container mx-auto py-6">
      <div className="mb-8">
        <div className="flex items-center gap-3 mb-2">
          <div className="p-2 bg-primary/10 rounded-lg">
            <Plug className="h-6 w-6 text-primary" />
          </div>
          <h1 className="text-3xl font-bold">Integrations</h1>
        </div>
        <p className="text-muted-foreground">
          Extend Anybase with powerful integrations and third-party services
        </p>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {integrations.map((integration) => {
          const Icon = integration.icon;
          const isAvailable = integration.status === 'available';
          
          return (
            <Card 
              key={integration.id} 
              className={`relative group transition-all duration-200 ${
                isAvailable 
                  ? 'hover:shadow-md hover:border-primary/20 cursor-pointer' 
                  : 'opacity-75'
              }`}
            >
              <CardHeader className="pb-3">
                <div className="flex items-start justify-between mb-3">
                  <div className={`p-2.5 rounded-lg ${
                    isAvailable 
                      ? 'bg-primary/10 group-hover:bg-primary/15' 
                      : 'bg-muted'
                  }`}>
                    <Icon className={`h-5 w-5 ${
                      isAvailable ? 'text-primary' : 'text-muted-foreground'
                    }`} />
                  </div>
                  {isAvailable && (
                    <Badge variant="outline" className="text-xs border-green-500/50 text-green-600 dark:text-green-400">
                      <Sparkles className="h-3 w-3 mr-1" />
                      Available
                    </Badge>
                  )}
                </div>
                <CardTitle className="text-lg font-semibold">
                  {integration.title}
                </CardTitle>
                <CardDescription className="text-sm line-clamp-2 mt-1">
                  {integration.description}
                </CardDescription>
              </CardHeader>
              
              <CardContent className="pt-0">
                {integration.features && (
                  <div className="flex flex-wrap gap-1 mb-4">
                    {integration.features.map((feature, idx) => (
                      <Badge 
                        key={idx} 
                        variant="secondary" 
                        className="text-xs font-normal"
                      >
                        {feature}
                      </Badge>
                    ))}
                  </div>
                )}
                
                <Link href={integration.href}>
                  <Button 
                    className="w-full group/btn" 
                    variant="outline"
                    size="sm"
                  >
                    <span>Configure</span>
                    <ArrowRight className="h-3.5 w-3.5 ml-1 transition-transform group-hover/btn:translate-x-0.5" />
                  </Button>
                </Link>
              </CardContent>
            </Card>
          );
        })}
      </div>
    </div>
  );
}