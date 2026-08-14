import { ComponentFixture, TestBed } from '@angular/core/testing';
import { HttpClientTestingModule } from '@angular/common/http/testing';
import { AdminDashboardComponent } from './admin.component';

describe('AdminDashboardComponent (Dedicated Web Admin Dashboard Test)', () => {
  let component: AdminDashboardComponent;
  let fixture: ComponentFixture<AdminDashboardComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [AdminDashboardComponent, HttpClientTestingModule]
    }).compileComponents();

    fixture = TestBed.createComponent(AdminDashboardComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create the admin dashboard component', () => {
    expect(component).toBeTruthy();
  });

  it('should initialize with stores tab active', () => {
    expect(component.activeTab()).toBe('stores');
  });

  it('should switch active tabs correctly', () => {
    component.activeTab.set('couriers');
    expect(component.activeTab()).toBe('couriers');

    component.activeTab.set('tariffs');
    expect(component.activeTab()).toBe('tariffs');
  });

  it('should open and close modals on signal trigger', () => {
    expect(component.displayAddStoreModal()).toBeFalse();
    component.displayAddStoreModal.set(true);
    expect(component.displayAddStoreModal()).toBeTrue();
  });
});
