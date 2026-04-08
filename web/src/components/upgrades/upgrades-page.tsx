'use client';

import { PageCard, PageShell } from '@/components/layout/page-shell';
import { ReleaseTable } from '@/components/upgrades/release-table';
import { UpgradesPageHeader } from '@/components/upgrades/upgrades-page-header';
import type { ReleaseItem } from '@/types/upgrade';

export function UpgradesPage({ releases }: { releases: ReleaseItem[] }) {
  return (
    <PageShell>
      <UpgradesPageHeader />
      <PageCard>
        <ReleaseTable releases={releases} />
      </PageCard>
    </PageShell>
  );
}
