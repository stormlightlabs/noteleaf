import type {ReactNode} from 'react';
import clsx from 'clsx';
import Heading from '@theme/Heading';
import styles from './styles.module.css';

type FeatureItem = {
  title: string;
  description: ReactNode;
};

const FeatureList: FeatureItem[] = [
  {
    title: 'Terminal-First Interface',
    description: (
      <>
        Built for the command line with a beautiful, keyboard-driven interface
        using the Bubble Tea framework.
      </>
    ),
  },
  {
    title: 'Task Management',
    description: (
      <>
        Organize your tasks with projects, priorities, tags, and due dates.
        Track your progress with statuses and time estimates.
      </>
    ),
  },
  {
    title: 'Knowledge Base',
    description: (
      <>
        Keep notes, track books, movies, and TV shows. Link everything together
        with tags and projects.
      </>
    ),
  },
];

function Feature({title, description}: FeatureItem) {
  return (
    <div className={clsx('col col--4')}>
      <div className="text--center padding-horiz--md">
        <Heading as="h3">{title}</Heading>
        <p>{description}</p>
      </div>
    </div>
  );
}

export default function HomepageFeatures(): ReactNode {
  return (
    <section className={styles.features}>
      <div className="container">
        <div className="row">
          {FeatureList.map((props, idx) => (
            <Feature key={idx} {...props} />
          ))}
        </div>
      </div>
    </section>
  );
}
