import {syntaxTree} from '@codemirror/language';
import {Range} from '@codemirror/state';
import {EditorView, Decoration, WidgetType} from '@codemirror/view';
import {computePosition, flip, shift, offset, arrow} from '@floating-ui/dom';
import at from 'lodash/at';

function createUserAvatarSVGWidget() {
  return new (class extends WidgetType {
    toDOM() {
      const container = document.createElement('span');
      container.style.position = 'relative';
      container.style.width = '0';
      container.style.height = '0';

      // This is the SVG code for userAvatar SVG
      const svg = document.createElementNS('http://www.w3.org/2000/svg', 'svg');
      svg.setAttribute('width', '20');
      svg.setAttribute('height', '20');
      svg.setAttribute('viewBox', '0 0 20 20');
      svg.setAttribute('fill', 'current-color');
      svg.setAttribute('xmlns', 'http://www.w3.org/2000/svg');

      svg.style.position = 'absolute';
      svg.style.left = '-1.5em';
      svg.style.transform = 'scale(0.6)';
      svg.setAttribute(
        'data-testid',
        'dynamicEffectiveSpecDecorations__userAvatarSVG',
      );

      const path = document.createElementNS(
        'http://www.w3.org/2000/svg',
        'path',
      );
      path.setAttribute('fill-rule', 'evenodd');
      path.setAttribute('clip-rule', 'evenodd');
      path.setAttribute(
        'd',
        'M9.99999 0.833332C6.77833 0.833332 4.16666 3.445 4.16666 6.66667C4.16666 8.5547 5.06363 10.2332 6.45469 11.2994C5.63304 11.5417 4.87478 11.8853 4.2121 12.3456C2.62291 13.4494 1.66666 15.1693 1.66666 17.5V20H3.33332V17.5C3.33332 15.6885 4.04374 14.4918 5.16288 13.7145C6.32565 12.9068 8.01014 12.5 9.99999 12.5C11.992 12.5 13.6762 12.9015 14.8381 13.706C15.9553 14.4794 16.6667 15.6749 16.6667 17.5V20H18.3333V17.5C18.3333 15.1584 17.3781 13.4373 15.7868 12.3357C15.1254 11.8777 14.3688 11.5367 13.5493 11.2963C14.9381 10.23 15.8333 8.55287 15.8333 6.66667C15.8333 3.445 13.2217 0.833332 9.99999 0.833332ZM9.99999 10.8333C12.3012 10.8333 14.1667 8.96785 14.1667 6.66667C14.1667 4.36548 12.3012 2.5 9.99999 2.5C7.6988 2.5 5.83332 4.36548 5.83332 6.66667C5.83332 8.96785 7.6988 10.8333 9.99999 10.8333Z',
      );

      svg.appendChild(path);
      container.append(svg);

      const tooltip = document.createElement('div');
      tooltip.setAttribute('id', 'tooltip');
      tooltip.setAttribute('role', 'tooltip');
      tooltip.setAttribute(
        'data-testid',
        'dynamicEffectiveSpecDecorations__tooltip',
      );
      tooltip.textContent =
        'User provided values that\nmay have overwritten a default';

      tooltip.style.display = 'none';
      tooltip.style.width = 'max-content';
      tooltip.style.position = 'absolute';
      tooltip.style.top = '0';
      tooltip.style.left = '0';
      tooltip.style.background = '#222';
      tooltip.style.color = 'white';
      tooltip.style.fontWeight = 'bold';
      tooltip.style.padding = '8px 16px';
      tooltip.style.borderRadius = '4px';
      tooltip.style.fontSize = '90%';

      const arrowElement = document.createElement('div');
      arrowElement.setAttribute('id', 'arrow');
      arrowElement.style.position = 'absolute';
      arrowElement.style.background = '#222';
      arrowElement.style.width = '8px';
      arrowElement.style.height = '8px';
      arrowElement.style.transform = 'rotate(45deg)';

      tooltip.append(arrowElement);

      function update() {
        computePosition(svg, tooltip, {
          placement: 'right',
          middleware: [
            offset(6),
            flip(),
            shift({padding: 5}),
            arrow({element: arrowElement}),
          ],
        }).then(({x, y, placement, middlewareData}) => {
          Object.assign(tooltip.style, {
            left: `${x}px`,
            top: `${y}px`,
          });

          // Accessing the data
          const {x: arrowX, y: arrowY} = middlewareData.arrow as any;

          const staticSide = {
            top: 'bottom',
            right: 'left',
            bottom: 'top',
            left: 'right',
          }[placement.split('-')[0]];

          Object.assign(arrowElement.style, {
            left: arrowX != null ? `${arrowX}px` : '',
            top: arrowY != null ? `${arrowY}px` : '',
            right: '',
            bottom: '',
          });

          if (staticSide) {
            Object.assign(arrowElement.style, {
              [staticSide]: '-4px',
            });
          }
        });
      }

      function showTooltip() {
        tooltip.style.display = 'block';
        update();
      }

      function hideTooltip() {
        tooltip.style.display = 'none';
      }

      svg.addEventListener('mouseenter', showTooltip);
      svg.addEventListener('mouseleave', hideTooltip);
      svg.addEventListener('focus', showTooltip);
      svg.addEventListener('blur', hideTooltip);

      container.append(tooltip);

      return container;
    }
  })();
}
function createStrikethroughWidget(text: string) {
  return new (class extends WidgetType {
    toDOM() {
      const span = document.createElement('span');
      span.textContent = ' ';
      const strikeThrough = document.createElement('s');

      // https://stackoverflow.com/questions/68787871/how-to-treat-text-with-special-characters-as-raw-literal-text-in-textcontent-or
      // \n \t \b will be rendered as non-literal characters otherwise
      const str = JSON.stringify(text);
      const cutStr = str.slice(1, str.length - 1);

      strikeThrough.textContent = cutStr;
      span.appendChild(strikeThrough);
      span.style.color = 'grey';

      strikeThrough.setAttribute('data-testid', 'overWrittenValue');
      return span;
    }
  })();
}

export const dynamicEffectiveSpecDecorations = (
  userOverrides: Record<string, unknown>,
) =>
  EditorView.decorations.compute([], (state) => {
    const decorations: Range<Decoration>[] = [];
    const keyStack: string[] = [];
    let prevName: string | undefined;
    let prevValue: number | string | boolean | undefined;

    syntaxTree(state).iterate({
      enter: (node) => {
        const value = state.sliceDoc(node.from, node.to);
        const lineNumber = state.doc.lineAt(node.from);
        // console.log(
        //   node.type.name === 'Property' || node.type.name === 'JsonText'
        //     ? '❌'
        //     : '',
        //   `line ${lineNumber.number} --`,
        //   node.type.name,
        //   '->',
        //   value,
        // );

        const valueWithoutQuotes = value.slice(1, -1);

        // If I am the first key inside of an object, add my parent's key to the stack.
        if (node.type.name === 'Object' && prevName === 'PropertyName') {
          keyStack.push(String(prevValue));
        }

        // Arrays are edge case. We do not want to travese into an array.
        const isArray = node.type.name === 'Array';

        // Create a decoration if I am at a leaf
        if (
          prevName === 'PropertyName' &&
          ['String', 'Number', 'Boolean', 'True', 'False', 'Array'].includes(
            node.type.name,
          )
        ) {
          keyStack.push(String(prevValue));
          // console.log(
          //   `🍃 line ${lineNumber.number} --`,
          //   stack.join('.'),
          //   '->',
          //   value,
          //   '-',
          //   'This',
          //   isArray ? 'IS' : 'not',
          //   'an array.',
          // );

          // Display the icon and previous value
          const value = at(userOverrides, keyStack.join('.'))[0];

          if (typeof value !== 'undefined') {
            decorations.push(
              Decoration.widget({
                widget: createUserAvatarSVGWidget(),
              }).range(lineNumber.from),
            );

            if (value !== null)
              decorations.push(
                Decoration.widget({
                  widget: createStrikethroughWidget(String(value)),
                }).range(lineNumber.to),
              );
          }
        }

        prevName = node.type.name;
        prevValue = valueWithoutQuotes;

        return !isArray;
      },
      leave: (node) => {
        // console.log(node.name === 'Property' ? '❗🏃' : '🏃', node.name, {
        //   value: view.state.sliceDoc(node.from, node.to),
        // console.log({stack: [...this.stack]});

        if (node.type.name === 'Property') {
          keyStack.pop();
        }
      },
    });

    return Decoration.set(decorations);
  });
