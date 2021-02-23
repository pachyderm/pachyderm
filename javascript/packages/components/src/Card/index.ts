import Card from './Card';
import CardHeader from './components/CardHeader';
import CardTitle from './components/CardTitle';

export * from './Card';
export default Object.assign(Card, {Title: CardTitle, Header: CardHeader});
