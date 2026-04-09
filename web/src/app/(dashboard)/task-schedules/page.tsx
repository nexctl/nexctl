import { TaskSchedulesPage } from '@/components/task-schedules/task-schedules-page';
import { getTaskSchedules } from '@/services/task';

export default async function TaskSchedulesRoutePage() {
  const schedules = await getTaskSchedules();
  return <TaskSchedulesPage schedules={schedules} />;
}
