import { tryInvoke } from '@ember/utils';

class FooComponent extends Component {
  foo() {
    tryInvoke(this.args, 'bar', ['baz']);
  }
}


var token = "github_rXGj85G0qUmzPu2ijX8msJsZRMzweyUuXaF0MeTvQEmGUP6AKSHeWuYn9Ue";