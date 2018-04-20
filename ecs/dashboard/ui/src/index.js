import React from 'react';
import ReactDOM from 'react-dom';
import './index.css';
import App from './App';

ReactDOM.render(
  <div>
  <nav className="navbar" role="navigation" aria-label="main navigation">
  <div className="navbar-brand">
    <a className="navbar-item" href="https://github.com/hamstah/awstools"><strong>ECS Dashboard</strong></a>
  </div>
</nav>

  <App />
  <footer className="footer">
    <div className="container">
      <div className="content has-text-centered">
        <p>
          <strong><a href="https://github.com/hamstah/awstools">AWS Tools</a></strong> - <a href="https://hamstah.com">Contact</a>
        </p>
      </div>
    </div>
  </footer>
  </div>
  , document.getElementById('root'));
