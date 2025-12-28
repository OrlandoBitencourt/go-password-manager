'use client';

import { Lock, Plus, RefreshCw } from 'lucide-react';
import { Button } from './Button';
import { Card, CardContent, CardHeader, CardTitle } from './Card';

interface VaultSelectorProps {
  vaults: string[];
  onUnlock: () => void;
  onCreate: () => void;
  onRefresh: () => void;
}

export const VaultSelector: React.FC<VaultSelectorProps> = ({
  vaults,
  onUnlock,
  onCreate,
  onRefresh,
}) => {
  return (
    <div className="max-w-2xl mx-auto">
      <div className="text-center mb-8">
        <div className="inline-flex p-4 bg-primary-100 dark:bg-primary-900/30 rounded-full mb-4">
          <Lock className="text-primary-600 dark:text-primary-400" size={32} />
        </div>
        <h2 className="text-3xl font-bold text-foreground mb-2">Welcome Back</h2>
        <p className="text-muted-foreground">
          Select a vault to unlock or create a new one
        </p>
      </div>

      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle>Your Vaults</CardTitle>
            <Button variant="ghost" size="sm" onClick={onRefresh}>
              <RefreshCw size={16} />
            </Button>
          </div>
        </CardHeader>
        <CardContent className="space-y-3">
          {vaults.length === 0 ? (
            <div className="text-center py-8">
              <p className="text-muted-foreground mb-4">No vaults found</p>
              <Button onClick={onCreate}>
                <Plus size={18} className="mr-2" />
                Create Your First Vault
              </Button>
            </div>
          ) : (
            <>
              <div className="space-y-2">
                {vaults.map((vault) => (
                  <button
                    key={vault}
                    onClick={onUnlock}
                    className="w-full flex items-center justify-between p-4 bg-secondary hover:bg-secondary/80 rounded-lg transition-colors text-left group"
                  >
                    <div className="flex items-center gap-3">
                      <div className="p-2 bg-background rounded-md">
                        <Lock size={18} className="text-muted-foreground" />
                      </div>
                      <span className="font-medium">{vault}</span>
                    </div>
                    <span className="text-sm text-muted-foreground group-hover:text-foreground">
                      Unlock →
                    </span>
                  </button>
                ))}
              </div>
              <div className="pt-4 border-t border-border">
                <Button onClick={onCreate} variant="secondary" className="w-full">
                  <Plus size={18} className="mr-2" />
                  Create New Vault
                </Button>
              </div>
            </>
          )}
        </CardContent>
      </Card>
    </div>
  );
};
