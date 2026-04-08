import { TasksPage } from '@/components/tasks/tasks-page';
import { getTasks } from '@/services/task';

export default async function TasksRoutePage() {
  const tasks = await getTasks();
  return <TasksPage tasks={tasks} />;
}
