import { HttpClient } from '@angular/common/http';
import { Component } from '@angular/core';
import { Router, RouterLink } from '@angular/router';

@Component
  ({
    selector: 'app-login',
    templateUrl: './login.component.html',
    styleUrls: ['./login.component.css']
  })

export class LoginComponent {
  constructor(public router: Router, private http: HttpClient) { }

  // Input fields
  courseID: string | undefined;
  username: string | undefined;
  password: string | undefined;
  confirmPassword: string | undefined;

  student() {
    if (this.courseID == "admin") {
      this.router.navigate(['student-view']);
    }
  }

  teacher() {
    if (this.username == "admin") {
      this.router.navigate(['teacher-view']);
    }
  }

  register(credentials: { username: string, password: string }) {
    if (this.password == this.confirmPassword) {
      /*OLD TEST
      console.log(credentials);
      this.http.post('localhost:3306/users.json', credentials).subscribe((res) =>
      {
        console.log(res);
      })
      */

      this.http.post('http://localhost:8080/register', {
        firstName: "Test_F",
        lastName: "Test_L",
        email: this.username,
        password: this.password,
      }).subscribe((response: any) => {
        if (response) {
          localStorage.setItem('token', response.jwt)
          this.router.navigate(['profile'])
        }
      })
    }
  }
}
