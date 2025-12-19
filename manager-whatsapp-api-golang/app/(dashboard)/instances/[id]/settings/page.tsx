'use client';

import { use } from 'react';
import { useRouter } from 'next/navigation';
import { ArrowLeft } from 'lucide-react';

import { useInstance } from '@/hooks';
import { InstanceSettingsForm } from '@/components/instances';
import { PageHeader } from '@/components/shared/page-header';
import { Button } from '@/components/ui/button';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { Skeleton } from '@/components/ui/skeleton';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';

interface SettingsPageProps {
  params: Promise<{ id: string }>;
}

export default function SettingsPage({ params }: SettingsPageProps) {
  const resolvedParams = use(params);
  const router = useRouter();
  const { instance, isLoading, error } = useInstance(resolvedParams.id);

  if (error) {
    return (
      <div className="space-y-6">
        <Button
          variant="ghost"
          onClick={() => router.back()}
          className="mb-4"
        >
          <ArrowLeft className="mr-2 h-4 w-4" />
          Voltar
        </Button>
        <Alert variant="destructive">
          <AlertTitle>Erro ao carregar instancia</AlertTitle>
          <AlertDescription>
            {error.message || 'Nao foi possivel carregar as informacoes da instancia.'}
          </AlertDescription>
        </Alert>
      </div>
    );
  }

  if (isLoading || !instance) {
    return (
      <div className="space-y-6">
        <Skeleton className="h-8 w-64" />
        <Skeleton className="h-96 w-full" />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <Button
          variant="ghost"
          onClick={() => router.back()}
          size="icon"
        >
          <ArrowLeft className="h-4 w-4" />
        </Button>
        <PageHeader
          title="Configuracoes da Instancia"
          description={`Configure o comportamento da instancia ${instance.name}`}
        />
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Configuracoes de Mensagens e Chamadas</CardTitle>
          <CardDescription>
            Personalize como sua instancia lida com mensagens e chamadas recebidas.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <InstanceSettingsForm
            instanceId={instance.id}
            instanceToken={instance.instanceToken}
            initialValues={{
              autoReadMessage: instance.autoReadMessage || false,
              callRejectAuto: instance.callRejectAuto || false,
              callRejectMessage: instance.callRejectMessage || '',
            }}
          />
        </CardContent>
      </Card>
    </div>
  );
}
