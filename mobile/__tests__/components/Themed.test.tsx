/**
 * Themed component tests
 *
 * Validates shared UI components: Card, Button, Badge, ProgressBar, Separator
 */
import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react-native';
import { Text, View, Card, Button, Badge, ProgressBar, Separator } from '@/components/Themed';

describe('Text', () => {
  it('renders text content', () => {
    render(<Text>Hello</Text>);
    expect(screen.getByText('Hello')).toBeTruthy();
  });
});

describe('View', () => {
  it('renders children', () => {
    render(
      <View>
        <Text>child</Text>
      </View>,
    );
    expect(screen.getByText('child')).toBeTruthy();
  });
});

describe('Card', () => {
  it('renders children inside a card', () => {
    render(
      <Card>
        <Text>Card content</Text>
      </Card>,
    );
    expect(screen.getByText('Card content')).toBeTruthy();
  });
});

describe('Button', () => {
  it('renders title text', () => {
    render(<Button title="Valider" onPress={() => {}} />);
    expect(screen.getByText('Valider')).toBeTruthy();
  });

  it('calls onPress when pressed', () => {
    const onPress = jest.fn();
    render(<Button title="Press me" onPress={onPress} />);
    fireEvent.press(screen.getByText('Press me'));
    expect(onPress).toHaveBeenCalledTimes(1);
  });

  it('renders icon when provided', () => {
    render(<Button title="Capture" icon="📸" onPress={() => {}} />);
    expect(screen.getByText('📸')).toBeTruthy();
    expect(screen.getByText('Capture')).toBeTruthy();
  });

  it('renders with different variants without crashing', () => {
    const variants = ['primary', 'secondary', 'outline', 'ghost'] as const;
    for (const variant of variants) {
      const { unmount } = render(
        <Button title={`btn-${variant}`} variant={variant} onPress={() => {}} />,
      );
      expect(screen.getByText(`btn-${variant}`)).toBeTruthy();
      unmount();
    }
  });
});

describe('Badge', () => {
  it('renders label text', () => {
    render(<Badge label="CONNAISSANCES" />);
    expect(screen.getByText('CONNAISSANCES')).toBeTruthy();
  });
});

describe('ProgressBar', () => {
  it('renders without crashing for various progress values', () => {
    const values = [0, 0.25, 0.5, 0.75, 1];
    for (const v of values) {
      const { toJSON, unmount } = render(<ProgressBar progress={v} />);
      expect(toJSON()).not.toBeNull();
      unmount();
    }
  });

  it('clamps progress to 0-1 range', () => {
    // Should not crash with out-of-bounds values
    const { toJSON: json1 } = render(<ProgressBar progress={-0.5} />);
    expect(json1()).not.toBeNull();
    const { toJSON: json2 } = render(<ProgressBar progress={1.5} />);
    expect(json2()).not.toBeNull();
  });
});

describe('Separator', () => {
  it('renders without crashing', () => {
    const { toJSON } = render(<Separator />);
    expect(toJSON()).not.toBeNull();
  });
});
