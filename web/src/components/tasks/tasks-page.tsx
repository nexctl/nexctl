'use client';

import { useRouter } from 'next/navigation';
import { useState } from 'react';
import { PageCard, PageShell } from '@/components/layout/page-shell';
import { NewTaskModal } from '@/components/tasks/new-task-modal';
import { TaskFilter } from '@/components/tasks/task-filter';
import { TaskTable } from '@/components/tasks/task-table';
import { TasksPageHeader } from '@/components/tasks/tasks-page-header';
import type { TaskItem } from '@/types/task';

export function TasksPage({ tasks }: { tasks: TaskItem[] }) {
  const router = useRouter();
  const [newTaskOpen, setNewTaskOpen] = useState(false);

  return (
    <PageShell>
      <TasksPageHeader onNewTask={() => setNewTaskOpen(true)} />
      <PageCard>
        <TaskFilter />
      </PageCard>
      <PageCard>
        <TaskTable tasks={tasks} />
      </PageCard>
      <NewTaskModal
        open={newTaskOpen}
        onClose={() => setNewTaskOpen(false)}
        onCreated={() => router.refresh()}
      />
    </PageShell>
  );
}
