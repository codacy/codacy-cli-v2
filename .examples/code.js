import { tryInvoke } from '@ember/utils';

class FooComponent extends Component {
  foo() {
    tryInvoke(this.args, 'bar', ['baz']);
  }
}
