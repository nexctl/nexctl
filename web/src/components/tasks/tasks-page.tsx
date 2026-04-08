'use client';

import { PageCard, PageShell } from '@/components/layout/page-shell';
import { TaskFilter } from '@/components/tasks/task-filter';
import { TaskTable } from '@/components/tasks/task-table';
import { TasksPageHeader } from '@/components/tasks/tasks-page-header';
import type { TaskItem } from '@/types/task';

export function TasksPage({ tasks }: { tasks: TaskItem[] }) {
  return (
    <PageShell>
      <TasksPageHeader />
      <PageCard>
        <TaskFilter />
      </PageCard>
      <PageCard>
        <TaskTable tasks={tasks} />
      </PageCard>
    </PageShell>
  );
}
