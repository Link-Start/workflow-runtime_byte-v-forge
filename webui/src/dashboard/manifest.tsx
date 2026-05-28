import { Workflow } from 'lucide-react';
import { DashboardNavSection, type DashboardModuleRegistration } from '@byte-v-forge/common-ui';
import { WorkflowPage } from './workflow-page';
import './styles.css';

const registration: DashboardModuleRegistration = {
  manifest: {
    id: 'workflow-runtime',
    nav: [
      {
        key: 'workflow',
        label: 'Workflow',
        icon: 'workflow',
        section: DashboardNavSection.DASHBOARD_NAV_SECTION_INFRASTRUCTURE,
        required_services: ['workflow-runtime'],
        order: 90
      }
    ]
  },
  icons: {
    workflow: <Workflow size={17} />
  },
  views: {
    workflow: () => <WorkflowPage />
  }
};

export default registration;
