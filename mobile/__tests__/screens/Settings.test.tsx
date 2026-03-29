/**
 * Settings screen tests
 *
 * Validates:
 * - Z8-AC07: Parent invitation accessible in settings
 * - Basic settings rows render correctly
 */
import React from 'react';
import { render, screen } from '@testing-library/react-native';

jest.mock('react-native-safe-area-context', () => ({
  useSafeAreaInsets: () => ({ top: 44, bottom: 34, left: 0, right: 0 }),
}));

import SettingsScreen from '@/app/(tabs)/settings';

describe('Settings Screen', () => {
  it('shows profile row', () => {
    render(<SettingsScreen />);
    expect(screen.getByText('Profil')).toBeTruthy();
    expect(screen.getByText(/Hugo/)).toBeTruthy();
  });

  it('shows schedule row', () => {
    render(<SettingsScreen />);
    expect(screen.getByText('Emploi du temps')).toBeTruthy();
  });

  it('Z8-AC07: shows parent invitation row', () => {
    render(<SettingsScreen />);
    expect(screen.getByText('Inviter un parent')).toBeTruthy();
    expect(screen.getByText('Non lié')).toBeTruthy();
  });

  it('shows notifications row', () => {
    render(<SettingsScreen />);
    expect(screen.getByText('Notifications')).toBeTruthy();
  });

  it('shows export data row', () => {
    render(<SettingsScreen />);
    expect(screen.getByText('Exporter mes données')).toBeTruthy();
  });

  it('shows help & about rows', () => {
    render(<SettingsScreen />);
    expect(screen.getByText('Aide & feedback')).toBeTruthy();
    expect(screen.getByText('À propos')).toBeTruthy();
    expect(screen.getByText('v0.1.0')).toBeTruthy();
  });

  it('shows delete account button', () => {
    render(<SettingsScreen />);
    expect(screen.getByText('Supprimer mon compte')).toBeTruthy();
  });
});
