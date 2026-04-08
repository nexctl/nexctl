'use client';

import { PageCard, PageShell } from '@/components/layout/page-shell';
import { FileTable } from '@/components/files/file-table';
import { FilesPageHeader } from '@/components/files/files-page-header';
import type { FileItem } from '@/types/file';

export function FilesPage({ files }: { files: FileItem[] }) {
  return (
    <PageShell>
      <FilesPageHeader />
      <PageCard>
        <FileTable files={files} />
      </PageCard>
    </PageShell>
  );
}
